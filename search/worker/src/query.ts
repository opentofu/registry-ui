import { Client } from '@neondatabase/serverless';

const searchQuery = `
WITH search_terms AS (
  SELECT unnest(regexp_split_to_array($1, '[ /]+')) AS term
),
matched_entities AS (
  SELECT *,
    -- Count the number of search terms that match across title, description, and link_variables->>'name'
    (
      SELECT COUNT(*)
      FROM search_terms st
      WHERE title ILIKE '%' || st.term || '%'
        OR description ILIKE '%' || st.term || '%'
        OR link_variables->>'name' ILIKE '%' || st.term || '%'
    ) AS term_match_count
  FROM entities
  WHERE EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE id ILIKE '%' || st.term || '%'
        OR addr ILIKE '%' || st.term || '%'
        OR title ILIKE '%' || st.term || '%'
        OR description ILIKE '%' || st.term || '%'
        OR link_variables->>'name' ILIKE '%' || st.term || '%'
    )
)
SELECT *,
  CASE
    -- Exact match for the entire search phrase
    WHEN title = $1 THEN 1
    WHEN link_variables->>'name' = $1 THEN 2
    WHEN description = $1 THEN 3 -- Exact match for individual words at the start
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE title ILIKE st.term || '%'
    ) THEN 4
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE link_variables->>'name' ILIKE st.term || '%'
    ) THEN 5
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE description ILIKE st.term || '%'
    ) THEN 6 -- Partial match for individual words
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE title ILIKE '%' || st.term || '%'
    ) THEN 7
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE link_variables->>'name' ILIKE '%' || st.term || '%'
    ) THEN 8
    WHEN EXISTS (
      SELECT 1
      FROM search_terms st
      WHERE description ILIKE '%' || st.term || '%'
    ) THEN 9
    ELSE 10
  END AS rank
FROM matched_entities
ORDER BY term_match_count DESC,
  -- Prioritize rows with more matching terms
  rank,
  LENGTH(title),
  LENGTH(link_variables->>'name'),
  LENGTH(description)
LIMIT 20;
`;

type Entity = {
	id: string;
	addr: string;
	title: string;
	description: string;
	link_variables: Record<string, string>;
};

export const query = async (client: Client, queryParam: string): Promise<Entity[]> => {
	const { rows } = await client.query(searchQuery, [queryParam]);
	return rows;
};
