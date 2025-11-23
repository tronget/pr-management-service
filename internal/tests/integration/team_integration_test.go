package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tronget/pr-management-service/internal/server"
	"github.com/tronget/pr-management-service/internal/tests/integration/testsuite"
)

func TestTeamEndpoints(t *testing.T) {
	pool := dbPool(t)
	suite := testsuite.New(t, pool, server.Handler(pool))
	t.Cleanup(suite.Close)

	t.Run("create team", func(t *testing.T) {
		payload := map[string]any{
			"team_name": "team-gamma",
			"members": []map[string]any{{
				"user_id":   "gamma-user-1",
				"username":  "Gary Gamma",
				"is_active": true,
			}},
		}
		res := suite.PostJSON(t, "/team/add", payload)
		require.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("get team not found", func(t *testing.T) {
		res := suite.Get(t, "/team/get?team_name=not-exist")
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}
