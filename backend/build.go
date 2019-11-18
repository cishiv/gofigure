package main

/**
	In the spirit of seperation of concerns, lets put our struct stuff here
**/

type Build struct {
	BuildID string `json:"buildID"`
	Time string `json:"time"`
	Action string `json:"action"`
	Outcome string `json:"outcome"`
}

func GetBuildByIDIndex(buildID string) int {
	for i,b := range buildHistory {
		if(b.BuildID == buildID) {
			return i
		}
	}
	return -1
}
func (build Build) SetOutcome(result string) {
	build.Outcome = result;
}
