package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("slug").NotEmpty().MaxLen(63).Unique(),
		field.String("host").NotEmpty().MaxLen(253).Unique(),
		field.String("name").NotEmpty().MaxLen(120),
		field.String("currency").MaxLen(3).Default("COP"),
		field.String("tax_name").Default("IVA"),
		field.String("timezone").Default("America/Bogota"),
		field.JSON("opening_hours", map[string]any{}).Default(map[string]any{}),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("canaries", IsolationCanary.Type),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("host").Unique(),
	}
}
