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
			 SELECT e.*
			 FROM entities e
					  INNER JOIN search_terms st
								 ON e.addr ILIKE '%' || st.term || '%'
		OR e.description ILIKE '%' || st.term || '%'
	GROUP BY e.id
		),
		ranked_entities AS (
	SELECT *,
		/* TODO: remove hard-coded hashicorp/opentofu preferential treatment */
		CASE WHEN link_variables->>'namespace' = 'hashicorp' THEN 1 WHEN link_variables->>'namespace' = 'opentofu' THEN 0 ELSE 0.5 END AS popularity_fudge,
		CASE WHEN type = 'provider' THEN 1 ELSE 0 END AS provider_rank_fudge,
		similarity(tm.addr, $1) AS title_sim,
		similarity(tm.description, $1) AS description_sim,
		similarity(link_variables->>'name', $1) AS name_sim
	FROM term_matches tm
		),
		providers AS (
	SELECT *
	FROM ranked_entities
	WHERE type LIKE 'provider%'
	ORDER BY
		provider_rank_fudge DESC,
		popularity_fudge DESC,
		title_sim DESC,
		name_sim DESC,
		description_sim DESC
		LIMIT 5
		),
		modules AS (
	SELECT *
	FROM ranked_entities
	WHERE type LIKE 'module%'
	ORDER BY
		popularity_fudge DESC,
		title_sim DESC,
		name_sim DESC,
		description_sim DESC
		LIMIT 5
		)
	SELECT * FROM providers
	UNION ALL
	SELECT * FROM modules;
`;
