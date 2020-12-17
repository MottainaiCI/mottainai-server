package client

import (
  setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
  "github.com/MottainaiCI/mottainai-server/routes/api/client/auth"
  "github.com/MottainaiCI/mottainai-server/routes/api/client/dashboard"
  "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
  m.Invoke(func(config *setting.Config) {
    m.Group(config.GetWeb().GroupAppPath(), func() {
      auth.Setup(m)
      dashboard.Setup(m)
    })
  })
}
