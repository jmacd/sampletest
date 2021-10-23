package kolmogorov

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKolmogorov(t *testing.T) {
	require.InEpsilon(t, 0.9999999996465492, K(10, 1), 1e-15)
	require.InEpsilon(t, 0.2512809600000001, K(10, 0.2), 1e-15)
	require.InEpsilon(t, 0.0467840289364274, K(100, 0.05), 1e-15)
}
