/** Cache lifetimes for page and endpoint responses, named by how fast the data moves. */
export const cache = {
	/** Search and facets, which change as soon as the index is reloaded. */
	search: 'public, max-age=0, s-maxage=600',
	/** Listings, where a new entry can wait an hour to appear. */
	listing: 'public, max-age=0, s-maxage=3600',
	/** A stored system or export, which only changes when that style is recrawled. */
	record: 'public, max-age=0, s-maxage=86400',
} as const;
