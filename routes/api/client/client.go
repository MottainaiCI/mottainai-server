package client

import (
  setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
  "github.com/MottainaiCI/mottainai-server/routes/api/client/auth"
  "gopkg.in/macaron.v1"

  "github.com/MottainaiCI/mottainai-server/routes/api/client/dashboard"

  _ "github.com/MottainaiCI/mottainai-server/client/statik"
  "github.com/rakyll/statik/fs"
)

func Setup(m *macaron.Macaron) {
  statikFS, err := fs.New()
  if err != nil {
    panic("fs")
  }

  // serves client
  m.Use(macaron.Static("",
    macaron.StaticOptions{
      FileSystem: statikFS,
    },
  ))

  m.Invoke(func(config *setting.Config) {
    m.Group(config.GetWeb().GroupAppPath(), func() {
      auth.Setup(m)
      dashboard.Setup(m)
    })
  })
}
