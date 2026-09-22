package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type IsolationCanary struct {
	ent.Schema
}

func (IsolationCanary) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("tenant_id", uuid.UUID{}),
		field.String("label").NotEmpty(),
	}
}

func (IsolationCanary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("canaries").
			Field("tenant_id").
			Required(),
	}
}

func (IsolationCanary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "label").Unique(),
	}
}
