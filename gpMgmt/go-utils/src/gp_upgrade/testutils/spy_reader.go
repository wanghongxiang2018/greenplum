package testutils

import "gp_upgrade/hub/configutils"

type SpyReader struct {
	Hostnames             []string
	MasterDataDir         string
	Port                  chan int
	SegmentConfigurations chan configutils.SegmentConfiguration

	Err error
}

func (r *SpyReader) GetMasterDataDir() string {
	return r.MasterDataDir
}

func (r *SpyReader) GetHostnames() ([]string, error) {
	return r.Hostnames, r.Err
}

func (r *SpyReader) GetSegmentConfiguration() configutils.SegmentConfiguration {
	var segmentConf configutils.SegmentConfiguration
	if len(r.SegmentConfigurations) != 0 {
		segmentConf = <-r.SegmentConfigurations
	}

	return segmentConf
}

func (r *SpyReader) OfOldClusterConfig(string) {}
func (r *SpyReader) OfNewClusterConfig(string) {}

func (r *SpyReader) GetPortForSegment(segmentDbid int) int {
	if len(r.Port) == 0 {
		return -1
	}

	return <-r.Port
}
