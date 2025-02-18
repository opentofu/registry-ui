import { DBClient, Entity } from './types';

export const query = async (client: DBClient, queryParam: string): Promise<Entity[]> => {
	const { rows } = await client.query(searchQuery, [queryParam]);
	return rows;
};

const searchQuery = `
	WITH search_terms AS (SELECT unnest(regexp_split_to_array($1, '[ /]+')) AS term),
		 term_matches AS (SELECT e.*
						  FROM entities e
								   INNER JOIN search_terms st
											  ON e.addr ILIKE '%' || st.term || '%'
												  OR e.description ILIKE '%' || st.term || '%'
						  GROUP BY id, last_updated, type, addr, version, title, description, link_variables, document, popularity, warnings),
		 ranked_entities AS (SELECT *,
								 /* The provider rank fudge ranks providers higher than their resources */
									CASE WHEN type = 'provider' THEN 1 ELSE 0 END          AS provider_rank_fudge,
								 /* When warnings are present, rank the provider lower because it's likely deprecated. */
									CASE WHEN warnings > 0 THEN -1 ELSE 0 END              AS warnings_rank_fudge,
								 /* Give a slight boost to providers with a higher star rating. */
								 /* Coalesce is used when popularity is zero, to avoid division by zero errors */
									tm.popularity / coalesce(NULLIF((SELECT max(popularity) FROM term_matches), 0), 1) AS popularity_rank,
								 /* Text similarity rankings, each taking a value from 0 to 1. */
									similarity(tm.addr, $1)                                AS title_sim,
									similarity(tm.description, $1)                         AS description_sim,
									similarity(link_variables ->> 'name', $1)              AS name_sim
							 FROM term_matches tm),
		 providers AS (SELECT *
					   FROM ranked_entities
					   WHERE type LIKE 'provider%'
					   ORDER BY (provider_rank_fudge + warnings_rank_fudge + 1) *(popularity_rank + title_sim + name_sim + description_sim/0.5) DESC
					   LIMIT 5),
		 modules AS (SELECT *
					 FROM ranked_entities
					 WHERE type LIKE 'module%'
					 ORDER BY (warnings_rank_fudge + 1) * (popularity_rank + title_sim + name_sim + description_sim/0.5) DESC
					 LIMIT 5)
	SELECT *
	FROM providers
	UNION ALL
	SELECT *
	FROM modules;
`;
