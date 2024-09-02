import { Client } from '@neondatabase/serverless';
import { Entity } from './types';

export const query = async (client: Client, queryParam: string): Promise<Entity[]> => {
	const { rows } = await client.query(searchQuery, [queryParam]);
	return rows;
};

const searchQuery = `
WITH search_terms AS (
  SELECT unnest(regexp_split_to_array($1, '[ /]+')) AS term
),
term_matches AS (
  SELECT e.*,
    COUNT(*) FILTER (WHERE e.title ILIKE '%' || st.term || '%'
                      OR e.description ILIKE '%' || st.term || '%'
                      OR e.link_variables->>'name' ILIKE '%' || st.term || '%') AS term_match_count
  FROM entities e
  CROSS JOIN search_terms st
  WHERE e.id ILIKE '%' || st.term || '%'
    OR e.addr ILIKE '%' || st.term || '%'
    OR e.title ILIKE '%' || st.term || '%'
    OR e.description ILIKE '%' || st.term || '%'
    OR e.link_variables->>'name' ILIKE '%' || st.term || '%'
  GROUP BY e.id
),
ranked_entities AS (
  SELECT *,
    CASE
      WHEN title = $1 THEN 1
      WHEN link_variables->>'name' = $1 THEN 2
      WHEN description = $1 THEN 3
      WHEN title ILIKE st.term || '%' THEN 4
      WHEN link_variables->>'name' ILIKE st.term || '%' THEN 5
      WHEN description ILIKE st.term || '%' THEN 6
      WHEN title ILIKE '%' || st.term || '%' THEN 7
      WHEN link_variables->>'name' ILIKE '%' || st.term || '%' THEN 8
      WHEN description ILIKE '%' || st.term || '%' THEN 9
      ELSE 10
    END AS rank
  FROM term_matches tm
  CROSS JOIN search_terms st
)
SELECT *
FROM ranked_entities
ORDER BY term_match_count DESC,
  rank,
  LENGTH(title),
  LENGTH(link_variables->>'name'),
  LENGTH(description)
LIMIT 20;
`;
