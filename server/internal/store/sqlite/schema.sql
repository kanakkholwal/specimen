PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 10000;

-- ---- crawl scheduling -------------------------------------------------------

CREATE TABLE IF NOT EXISTS frontier (
  url         TEXT PRIMARY KEY,
  kind        TEXT NOT NULL,
  depth       INTEGER NOT NULL DEFAULT 0,
  meta        TEXT,
  state       TEXT NOT NULL DEFAULT 'pending', -- pending | leased | done | failed
  attempts    INTEGER NOT NULL DEFAULT 0,
  status      INTEGER,
  error       TEXT,
  etag        TEXT,
  last_modified TEXT,
  fetched_at  TEXT,
  updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS frontier_state ON frontier(state, depth);

CREATE TABLE IF NOT EXISTS runs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  adapter     TEXT NOT NULL,
  started_at  TEXT NOT NULL,
  finished_at TEXT,
  fetched     INTEGER NOT NULL DEFAULT 0,
  parsed      INTEGER NOT NULL DEFAULT 0,
  not_modified INTEGER NOT NULL DEFAULT 0,
  failed      INTEGER NOT NULL DEFAULT 0,
  bytes       INTEGER NOT NULL DEFAULT 0,
  notes       TEXT
);

CREATE TABLE IF NOT EXISTS fetch_log (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id    INTEGER,
  url       TEXT NOT NULL,
  status    INTEGER,
  bytes     INTEGER,
  ms        INTEGER,
  attempts  INTEGER,
  error     TEXT,
  at        TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS fetch_log_url ON fetch_log(url);

-- ---- identity ---------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sites (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  origin     TEXT NOT NULL UNIQUE,   -- www stripped, subdomain kept
  etld1      TEXT NOT NULL,
  subdomain  TEXT NOT NULL DEFAULT '',
  site_name  TEXT,
  first_seen TEXT NOT NULL DEFAULT (datetime('now')),
  last_seen  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS sites_etld1 ON sites(etld1);

CREATE TABLE IF NOT EXISTS styles (
  id            TEXT PRIMARY KEY,          -- refero uuid
  site_id       INTEGER REFERENCES sites(id),
  url           TEXT,
  site_name     TEXT,
  source_url    TEXT,
  color_scheme  TEXT,
  theme         TEXT,
  industry      TEXT,
  north_star    TEXT,
  north_star_detail TEXT,
  description   TEXT,
  layout        TEXT,
  imagery       TEXT,
  is_dark_mode  INTEGER,
  colorfulness  REAL,
  density       TEXT,
  base_unit     REAL,
  scale_base    REAL,
  scale_name    TEXT,
  scale_ratio   REAL,
  scale_confidence REAL,
  viewport_w    INTEGER,
  viewport_h    INTEGER,
  element_count INTEGER,
  duration_ms   INTEGER,
  extracted_at  TEXT,
  created_at    TEXT,
  fetched_at    TEXT,
  result_sha256 TEXT,
  has_result    INTEGER NOT NULL DEFAULT 0,
  has_card      INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS styles_site ON styles(site_id);
CREATE INDEX IF NOT EXISTS styles_complete ON styles(has_result, has_card);

-- ---- shared value tables: one row per distinct value across the whole corpus --

CREATE TABLE IF NOT EXISTS colors (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  hex     TEXT NOT NULL UNIQUE,
  oklch_l REAL, oklch_c REAL, oklch_h REAL
);

CREATE TABLE IF NOT EXISTS fonts (
  id     INTEGER PRIMARY KEY AUTOINCREMENT,
  family TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  UNIQUE(family, source)
);

CREATE TABLE IF NOT EXISTS shadows  (id INTEGER PRIMARY KEY AUTOINCREMENT, value TEXT NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS radii    (id INTEGER PRIMARY KEY AUTOINCREMENT, value REAL NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS spacings (id INTEGER PRIMARY KEY AUTOINCREMENT, value REAL NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS gradients(
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  value TEXT NOT NULL UNIQUE,
  type  TEXT
);

-- ---- style to value edges ----------------------------------------------------

CREATE TABLE IF NOT EXISTS style_colors (
  style_id   TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  color_id   INTEGER NOT NULL REFERENCES colors(id),
  origin     TEXT NOT NULL,          -- raw | system | card | surface
  name       TEXT,
  role       TEXT,
  group_name TEXT,
  level      INTEGER,
  purpose    TEXT,
  gradient   TEXT,
  frequency  INTEGER,
  confidence REAL,
  prominence REAL,
  ord        INTEGER,
  PRIMARY KEY (style_id, color_id, origin)
);
CREATE INDEX IF NOT EXISTS style_colors_color ON style_colors(color_id);

CREATE TABLE IF NOT EXISTS color_usage (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  color_id INTEGER NOT NULL REFERENCES colors(id),
  context  TEXT NOT NULL,
  property TEXT NOT NULL,
  count    INTEGER NOT NULL,
  PRIMARY KEY (style_id, color_id, context, property)
);

CREATE TABLE IF NOT EXISTS contrast_pairs (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  fg_color_id INTEGER NOT NULL REFERENCES colors(id),
  bg_color_id INTEGER NOT NULL REFERENCES colors(id),
  ratio REAL,
  level TEXT,
  PRIMARY KEY (style_id, fg_color_id, bg_color_id)
);

CREATE TABLE IF NOT EXISTS style_fonts (
  style_id  TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  font_id   INTEGER NOT NULL REFERENCES fonts(id),
  weights   TEXT,
  frequency INTEGER,
  features  TEXT,
  PRIMARY KEY (style_id, font_id)
);
CREATE INDEX IF NOT EXISTS style_fonts_font ON style_fonts(font_id);

CREATE TABLE IF NOT EXISTS style_shadows (
  style_id  TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  shadow_id INTEGER NOT NULL REFERENCES shadows(id),
  contexts  TEXT,
  frequency INTEGER,
  PRIMARY KEY (style_id, shadow_id)
);

CREATE TABLE IF NOT EXISTS style_radii (
  style_id  TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  radius_id INTEGER NOT NULL REFERENCES radii(id),
  contexts  TEXT,
  frequency INTEGER,
  PRIMARY KEY (style_id, radius_id)
);

CREATE TABLE IF NOT EXISTS style_spacings (
  style_id   TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  spacing_id INTEGER NOT NULL REFERENCES spacings(id),
  contexts   TEXT,
  properties TEXT,
  pairs      TEXT,
  frequency  INTEGER,
  PRIMARY KEY (style_id, spacing_id)
);

CREATE TABLE IF NOT EXISTS style_gradients (
  style_id    TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  gradient_id INTEGER NOT NULL REFERENCES gradients(id),
  frequency   INTEGER,
  PRIMARY KEY (style_id, gradient_id)
);

-- ---- descriptive layer -------------------------------------------------------

CREATE TABLE IF NOT EXISTS type_steps (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  size REAL, family TEXT, weight INTEGER,
  line_height REAL, letter_spacing REAL, text_transform TEXT,
  contexts TEXT, frequency INTEGER,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS type_scale (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  role TEXT, size REAL, line_height REAL, letter_spacing REAL,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS typography_roles (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  role TEXT, family TEXT, sizes TEXT, weight TEXT, line_height TEXT,
  substitute TEXT, letter_spacing TEXT, font_features TEXT,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS elevation (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  element TEXT, style TEXT,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS components (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  name TEXT, role TEXT, description TEXT,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS guidelines (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,   -- do | dont
  ord INTEGER NOT NULL,
  text TEXT,
  PRIMARY KEY (style_id, kind, ord)
);

CREATE TABLE IF NOT EXISTS similar (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  business TEXT, why TEXT,
  PRIMARY KEY (style_id, ord)
);

CREATE TABLE IF NOT EXISTS spacing_map (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  key TEXT NOT NULL,     -- elementGap, sectionGap, radius.cards, ...
  value TEXT,
  PRIMARY KEY (style_id, key)
);

CREATE TABLE IF NOT EXISTS custom_sections (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord INTEGER NOT NULL,
  title TEXT, content TEXT,
  PRIMARY KEY (style_id, ord)
);

-- ---- media: every URL recorded, bytes optional -------------------------------

CREATE TABLE IF NOT EXISTS media (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  url         TEXT NOT NULL UNIQUE,
  kind        TEXT NOT NULL,   -- screenshot | thumbnail | icon | video | video_poster
  width       INTEGER,
  height      INTEGER,
  duration_ms INTEGER,
  sha256      TEXT,
  bytes       INTEGER,
  local_path  TEXT,
  downloaded  INTEGER NOT NULL DEFAULT 0,
  checked_at  TEXT
);
CREATE INDEX IF NOT EXISTS media_kind ON media(kind, downloaded);

CREATE TABLE IF NOT EXISTS style_media (
  style_id TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  media_id INTEGER NOT NULL REFERENCES media(id),
  role     TEXT NOT NULL,
  PRIMARY KEY (style_id, role)
);

-- ---- collections -------------------------------------------------------------

CREATE TABLE IF NOT EXISTS collections (
  id    INTEGER PRIMARY KEY AUTOINCREMENT,
  slug  TEXT NOT NULL UNIQUE,
  url   TEXT,
  title TEXT,
  kind  TEXT
);

CREATE TABLE IF NOT EXISTS collection_styles (
  collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  style_id      TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  ord           INTEGER,
  PRIMARY KEY (collection_id, style_id)
);

-- ---- derived artifacts -------------------------------------------------------

CREATE TABLE IF NOT EXISTS artifacts (
  style_id    TEXT NOT NULL REFERENCES styles(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  version     TEXT NOT NULL,
  sha256      TEXT,
  bytes       INTEGER,
  local_path  TEXT,
  rendered_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (style_id, name)
);
