package authorization

import (
	"context"
	"fmt"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthorizationService struct {
	logger      *logrus.Logger
	schema      string
	spiceClient *authzed.Client
	lastToken   *pb.ZedToken // Added to store the last ZedToken for consistency // TODO: store it in cache
	// TODO: add cache service
}

func NewAuthorizationService(
	logger *logrus.Logger,
	config *config.SpiceDBConfig,
	schema string,
) (*AuthorizationService, error) {
	// systemCerts, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
	// if err != nil {
	// 	logger.Fatalf("unable to load system CA certificates: %s", err)
	// }

	spicedbEndpoint := fmt.Sprintf("%s:%d", config.Host, config.Port)

	client, err := authzed.NewClient(
		spicedbEndpoint,

		// These options are if you're NOT self-hosting & want TLS:
		// grpcutil.WithBearerToken(config.PresharedKey),
		// grpc.WithTransportCredentials(insecure.NewCredentials()),

		// These options are if you're self-hosting and don't want TLS:
		grpcutil.WithInsecureBearerToken(config.PresharedKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %w", err)
	}

	return &AuthorizationService{
		logger:      logger,
		schema:      schema,
		spiceClient: client,
		lastToken:   nil, // Initialize lastToken as nil
	}, nil
}

func (s *AuthorizationService) WriteSchema() error {
	request := &pb.WriteSchemaRequest{
		Schema: s.schema,
	}
	_, err := s.spiceClient.WriteSchema(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	s.logger.Info("Schema written successfully")

	return nil
}

// WriteRelationships
func (s *AuthorizationService) WriteRelationships() error {
	request := &pb.WriteRelationshipsRequest{Updates: []*pb.RelationshipUpdate{
		{ // Emilia is a Writer on Post 1
			Operation: pb.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: string(PostResource),
					ObjectId:   string(Post1),
				},
				Relation: string(WriterRelation),
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: string(UserResource),
						ObjectId:   string(Emilia),
					},
				},
			},
		},
		{ // Beatrice is a Reader on Post 1
			Operation: pb.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: string(PostResource),
					ObjectId:   string(Post1),
				},
				Relation: string(ReaderRelation),
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: string(UserResource),
						ObjectId:   string(Beatrice),
					},
				},
			},
		},
	}}

	resp, err := s.spiceClient.WriteRelationships(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to write relations: %w", err)
	}

	s.lastToken = resp.WrittenAt
	s.logger.Info("Relationships written successfully, token: ", resp.WrittenAt.Token)

	return nil
}

// createSubjectReference creates a SubjectReference for a user
func (s *AuthorizationService) createSubjectReference(objectref *pb.ObjectReference) *pb.SubjectReference {
	return &pb.SubjectReference{
		Object: objectref,
	}
}

// createObjectReference creates an ObjectReference for a resource
func (s *AuthorizationService) createObjectReference(resourceType, resourceID string) *pb.ObjectReference {
	return &pb.ObjectReference{
		ObjectType: resourceType,
		ObjectId:   resourceID,
	}
}

// checkPermission checks if a subject has a specific permission on a resource
func (s *AuthorizationService) checkPermission(ctx context.Context, resource *pb.ObjectReference, permission string, subject *pb.SubjectReference) (bool, error) {
	req := &pb.CheckPermissionRequest{
		Resource:   resource,
		Permission: permission,
		Subject:    subject,
	}

	// Add consistency guarantee using the last ZedToken
	if s.lastToken != nil {
		req.Consistency = &pb.Consistency{
			Requirement: &pb.Consistency_AtLeastAsFresh{
				AtLeastAsFresh: s.lastToken,
			},
		}
	}

	resp, err := s.spiceClient.CheckPermission(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return resp.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

// verifyPermission checks a permission and logs an error if it doesn't match the expected value
func (s *AuthorizationService) verifyPermission(
	ctx context.Context,
	resource *pb.ObjectReference,
	permission Permission,
	subject *pb.SubjectReference,
) error {
	hasPermission, err := s.checkPermission(ctx, resource, string(permission), subject)
	if err != nil {
		return err
	}

	if hasPermission {
		s.logger.Infof("%s has %s permission on %s: %s", subject.Object.ObjectId, permission, resource.ObjectType, resource.ObjectId)
	} else {
		s.logger.Warnf("%s dont' have %s permission on %s: %s", subject.Object.ObjectId, permission, resource.ObjectType, resource.ObjectId)
	}
	return nil
}

// CheckPermissions verifies that the permissions are set up correctly
func (s *AuthorizationService) CheckPermissions() error {
	ctx := context.Background()

	// Create references
	emilia := s.createSubjectReference(
		s.createObjectReference(string(UserResource), string(Emilia)),
	)
	beatrice := s.createSubjectReference(
		s.createObjectReference(string(UserResource), string(Beatrice)),
	)
	firstPost := s.createObjectReference(string(PostResource), string(Post1))

	// Check Emilia's permissions
	if err := s.verifyPermission(ctx, firstPost, ReadPermission, emilia); err != nil {
		return err
	}
	if err := s.verifyPermission(ctx, firstPost, WritePermission, emilia); err != nil {
		return err
	}

	// Check Beatrice's permissions
	if err := s.verifyPermission(ctx, firstPost, ReadPermission, beatrice); err != nil {
		return err
	}
	if err := s.verifyPermission(ctx, firstPost, WritePermission, beatrice); err != nil {
		return err
	}

	return nil
}
