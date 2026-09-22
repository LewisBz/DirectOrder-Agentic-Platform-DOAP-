package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type GuestSession struct {
	ent.Schema
}

func (GuestSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("tenant_id", uuid.UUID{}),
		field.String("token_hash").NotEmpty().Sensitive().Unique(),
		field.String("channel").Default("pwa"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("last_seen").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (GuestSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("guest_sessions").Field("tenant_id").Required(),
	}
}

func (GuestSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("tenant_id"),
	}
}
