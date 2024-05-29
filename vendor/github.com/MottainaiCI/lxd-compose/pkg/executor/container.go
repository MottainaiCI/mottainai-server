/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	lxd_api "github.com/canonical/lxd/shared/api"
)

// Get the container data and the ETag
func (e *LxdCExecutor) GetContainer(name string) (*lxd_api.Container, string, error) {
	return e.LxdClient.GetContainer(name)
}

func (e *LxdCExecutor) UpdateContainer(name string, cdata *lxd_api.ContainerPut, etag string) error {
	oper, err := e.LxdClient.UpdateContainer(name, *cdata, etag)
	if err != nil {
		return err
	}

	err = e.WaitOperation(oper, nil)
	if err != nil {
		return err
	}

	e.Emitter.Emits(LxdContainerUpdated, map[string]interface{}{
		"name":      name,
		"profiles":  cdata.Profiles,
		"ephemeral": cdata.Ephemeral,
		"config":    cdata.Config,
		"devices":   cdata.Devices,
	})

	return nil
}

func (e *LxdCExecutor) RemoveProfilesFromContainer(name string, profiles []string) error {
	// Retrieve the current status of the container
	cdata, etag, err := e.GetContainer(name)
	if err != nil {
		return err
	}

	// Convert profiles to remove in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to remove are present
	newProfilesList := []string{}
	for _, p := range cdata.ContainerPut.Profiles {
		if _, present := mprofiles[p]; !present {
			newProfilesList = append(newProfilesList, p)
		}
	}

	cdata.ContainerPut.Profiles = newProfilesList

	err = e.UpdateContainer(name, &cdata.ContainerPut, etag)
	if err != nil {
		return err
	}

	return nil
}

func (e *LxdCExecutor) AddProfiles2Container(name string, profiles []string) error {
	// Retrieve the current status of the container
	cdata, etag, err := e.GetContainer(name)
	if err != nil {
		return err
	}

	// Convert profiles to add in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range cdata.ContainerPut.Profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to add are present
	update2do := false
	newProfilesList := cdata.ContainerPut.Profiles
	for _, p := range profiles {
		if _, present := mprofiles[p]; !present {
			update2do = true
			newProfilesList = append(newProfilesList, p)
		}
	}

	if update2do {
		cdata.ContainerPut.Profiles = newProfilesList

		err := e.UpdateContainer(name, &cdata.ContainerPut, etag)
		if err != nil {
			return err
		}
	}

	return nil
}
