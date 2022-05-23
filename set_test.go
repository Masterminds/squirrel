package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// select user_id from (
//   select user_id from user_identifiers
//   where
//     tenant_id='default' and
//     user_pool_id='pool' and
//     identifier_value like '%'
//   union select user_id from user_verifiable_addresses
//   where
//     tenant_id='default' and
//     user_pool_id='pool' and
//     verified_address like '%'
//   ) where
//     user_id > 'user1000009'
//   order by user_id
//   limit 10

func TestSetBuilder(t *testing.T) {
	set := NewSet(
		Select("user_id").
			From("user_identifiers").
			Where(Eq{
				"tenant_id":    "default",
				"user_pool_id": "pool",
			}).Where(Like{
			"identifier_value": "%",
		}),
	)

	set = set.Union(
		Select("user_id").
			From("user_verifiable_addresses").
			Where(Eq{
				"tenant_id":    "default",
				"user_pool_id": "pool",
			}).Where(Like{
			"identifier_value": "%",
		}))

	b := SelectFromSet(
		Select("u.user_id").
			Where(Gt{
				"user_id": "user100",
			}).OrderBy("u.user_id").
			Limit(10),
		set,
		"u")

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT u.user_id FROM (" +
		"SELECT user_id FROM user_identifiers WHERE tenant_id = ? AND user_pool_id = ? AND identifier_value LIKE ? " +
		"UNION ( SELECT user_id FROM user_verifiable_addresses WHERE tenant_id = ? AND user_pool_id = ? AND identifier_value LIKE ? ) " +
		") AS u WHERE user_id > ? ORDER BY u.user_id LIMIT 10"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"default", "pool", "%", "default", "pool", "%", "user100"}
	assert.Equal(t, expectedArgs, args)
}
