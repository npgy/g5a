package main

import "sync"

type GitHubAppState struct {
	mutex          sync.RWMutex
	appID          int64
	installationID int64
}

func (state *GitHubAppState) setAppID(appID int64) {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	state.appID = appID
}

func (state *GitHubAppState) setInstallationID(installationID int64) {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	state.installationID = installationID
}

func (state *GitHubAppState) ids() (int64, int64, bool) {
	state.mutex.RLock()
	defer state.mutex.RUnlock()

	if state.appID == 0 || state.installationID == 0 {
		return 0, 0, false
	}
	return state.appID, state.installationID, true
}
