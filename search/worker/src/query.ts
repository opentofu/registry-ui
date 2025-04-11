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
	     max_popularity AS (SELECT max(popularity) AS max_popularity FROM term_matches tm),
		 ranked_entities AS (SELECT *,
								 /* The rank fudge ranks providers and modules higher than their resources/submodules (excluding archived terraform-providers) */
									CASE WHEN (type = 'provider' OR type = 'module') AND addr NOT LIKE 'terraform-providers/%' AND addr NOT LIKE 'opentofu/%' THEN 1 ELSE 0 END          AS type_rank_fudge,
								 /* When warnings are present, rank the provider lower because it's likely deprecated. */
								 /* DISABLED CASE WHEN warnings > 1 THEN -1 ELSE 0 END              AS warnings_rank_fudge, */
								 0 as warnings_rank_fudge,
								 /* Give a slight boost to providers with a higher star rating. */
									tm.popularity / (SELECT CASE WHEN max_popularity > 0 THEN max_popularity ELSE 1 END FROM max_popularity) AS popularity_rank,
								 /* Text similarity rankings, each taking a value from 0 to 1. */
									similarity(tm.addr, $1)                                AS title_sim,
									similarity(tm.description, $1)                         AS description_sim,
									similarity(link_variables ->> 'name', $1)              AS name_sim
							 FROM term_matches tm),
		 providers AS (SELECT *
					   FROM ranked_entities
					   WHERE type LIKE 'provider%'
					   ORDER BY (type_rank_fudge + warnings_rank_fudge + 1) *(popularity_rank + title_sim + name_sim + description_sim/0.5) DESC
					   LIMIT 5),
		 modules AS (SELECT *
					 FROM ranked_entities
					 WHERE type LIKE 'module%'
					 ORDER BY (type_rank_fudge + warnings_rank_fudge + 1) * (popularity_rank + title_sim + name_sim + description_sim/0.5) DESC
					 LIMIT 5)
	SELECT *
	FROM providers
	UNION ALL
	SELECT *
	FROM modules;
`;
