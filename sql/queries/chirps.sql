-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: GetChirps :many
SELECT 
  * 
FROM chirps
WHERE
  CASE
    WHEN @author_id::TEXT IS NOT NULL AND @author_id::TEXT <> '' THEN user_id = (@author_id::UUID)
    ELSE TRUE
  END
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;