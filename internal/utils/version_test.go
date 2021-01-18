package utils

import "testing"

func TestNodeStatus_ComputerBuildVersionCode(t *testing.T) {
	{
		t.Log("", VersionToLong(""))
	}

	{
		t.Log("0.0.6", VersionToLong("0.0.6"))
	}

	{
		t.Log("0.0.6.1", VersionToLong("0.0.6.1"))
	}

	{
		t.Log("0.0.7", VersionToLong("0.0.7"))
	}

	{
		t.Log("0.7", VersionToLong("0.7"))
	}
	{
		t.Log("7", VersionToLong("7"))
	}
	{
		t.Log("7.0.1", VersionToLong("7.0.1"))
	}
}
