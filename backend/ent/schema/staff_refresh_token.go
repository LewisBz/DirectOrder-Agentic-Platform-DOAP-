package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type StaffRefreshToken struct {
	ent.Schema
}

func (StaffRefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("tenant_id", uuid.UUID{}),
		field.UUID("staff_id", uuid.UUID{}),
		field.String("token_hash").NotEmpty().Sensitive().Unique(),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (StaffRefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("refresh_tokens").Field("tenant_id").Required(),
		edge.From("staff", Staff.Type).Ref("refresh_tokens").Field("staff_id").Required(),
	}
}

func (StaffRefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("staff_id"),
	}
}
