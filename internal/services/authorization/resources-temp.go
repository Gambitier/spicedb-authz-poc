package authorization

type User string

const (
	Emilia   User = "emilia"
	Beatrice User = "beatrice"
)

type Post string

const (
	Post1 Post = "1"
)

type Resource string

const (
	PostResource Resource = "post"
	UserResource Resource = "user"
)

type Permission string

const (
	ReadPermission  Permission = "read"
	WritePermission Permission = "write"
)

type Relation string

const (
	ReaderRelation Relation = "reader"
	WriterRelation Relation = "writer"
)
