package ops

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfiles_Has(t *testing.T) {
	profiles := NewProfiles().Load()

	assert.True(t, profiles.Has(runtime.GOOS), "expected profile")
	assert.False(t, profiles.Has("missing"), "unexpected profile")
}

func TestProfiles_List(t *testing.T) {
	profiles := NewProfiles().Load()

	assert.Contains(t, profiles.List(), runtime.GOOS, "expected profile")
	assert.Contains(t, profiles.List(), runtime.GOARCH, "expected profile")
}

func TestProfiles_Load(t *testing.T) {
	t.Setenv("VALK_PROFILES", "ham,cheese")
	t.Setenv("KUBERNETES_SERVICE_HOST", "foo")
	t.Setenv("GOOGLE_COMPUTE_METADATA", "bar")

	profiles := NewProfiles().Load()

	assert.True(t, profiles.Has("ham"), "expected profile")
	assert.True(t, profiles.Has("cheese"), "expected profile")
	assert.True(t, profiles.Has("k8s"), "expected profile")
	assert.True(t, profiles.Has("gcp"), "expected profile")
}
