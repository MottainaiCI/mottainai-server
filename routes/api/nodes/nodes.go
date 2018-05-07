package nodesapi

import (
	nodes "github.com/MottainaiCI/mottainai-server/pkg/nodes"
	"github.com/go-macaron/binding"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	bind := binding.Bind
	m.Get("/api/nodes", ShowAll)
	m.Get("/api/nodes/add", APICreate)
	m.Get("/api/nodes/show/:id", Show)
	m.Get("/api/nodes/delete/:id", APIRemove)
	m.Get("/api/nodes/register", bind(nodes.Node{}), Register)
}
