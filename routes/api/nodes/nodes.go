package nodesapi

import (
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/pkg/nodes"
	"github.com/go-macaron/binding"
)

func Setup(m *mottainai.Mottainai) {
	bind := binding.Bind
	m.Get("/api/nodes", ShowAll)
	m.Get("/api/nodes/add", APICreate)
	m.Get("/api/nodes/show/:id", Show)
	m.Get("/api/nodes/delete/:id", APIRemove)
	m.Get("/api/nodes/register", bind(nodes.Node{}), Register)
}
