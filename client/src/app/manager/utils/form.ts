export function parseMapText(value: string): Record<string, string> {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .reduce((acc: Record<string, string>, pair) => {
      const [k, ...rest] = pair.split(":");
      const key = k?.trim();
      const mapped = rest.join(":").trim();
      if (key && mapped) acc[key] = mapped;
      return acc;
    }, {});
}

export function mapToText(value?: Record<string, string>): string {
  if (!value) return "";
  return Object.entries(value)
    .map(([k, v]) => `${k}:${v}`)
    .join(", ");
}

const INTEGER_FIELDS = new Set([
  "port",
  "batch_size",
  "flush_interval_ms",
  "max_retries",
  "retry_base_ms",
  "redis_ttl",
]);

const OPTIONAL_STRING_FIELDS = new Set([
  "username",
  "password",
  "slot_name",
  "publication_name",
  "topic",
  "name",
  "api_key",
  "index_prefix",
  "index",
]);

export function buildConfigPayloadFromFormData(
  formData: FormData,
  selectedSinkOrSourceType: string
): Record<string, unknown> {
  const data: Record<string, unknown> = {};

  formData.forEach((value, key) => {
    const raw = String(value).trim();
    if (key === "tables" || key === "redis_value_fields") {
      const list = raw.split(",").map((t) => t.trim()).filter(Boolean);
      if (key === "tables") data.tables = list;
      else data.redis_value_fields = list;
      return;
    }
    if (key === "url") {
      data.url = raw ? [raw] : [];
      return;
    }
    if (key === "index_mapping" || key === "field_mapping") {
      data[key] = parseMapText(raw);
      return;
    }
    if (INTEGER_FIELDS.has(key)) {
      if (!raw) return;
      const parsed = parseInt(raw, 10);
      if (!Number.isNaN(parsed)) data[key] = parsed;
      return;
    }
    if (OPTIONAL_STRING_FIELDS.has(key) && !raw) return;
    data[key] = raw;
  });

  if (selectedSinkOrSourceType === "redis") {
    const redis: Record<string, unknown> = {};
    if (data.redis_command) redis.command = data.redis_command;
    if (data.redis_key_template) redis.key_template = data.redis_key_template;
    if (Array.isArray(data.redis_value_fields) && (data.redis_value_fields as string[]).length > 0) {
      redis.value_fields = data.redis_value_fields;
    }
    if (typeof data.redis_ttl === "number") redis.ttl = data.redis_ttl;
    if (Object.keys(redis).length > 0) data.redis = redis;
  }

  delete data.redis_command;
  delete data.redis_key_template;
  delete data.redis_value_fields;
  delete data.redis_ttl;

  const im = data.index_mapping as Record<string, string> | undefined;
  if (im && Object.keys(im).length === 0) delete data.index_mapping;
  const fm = data.field_mapping as Record<string, string> | undefined;
  if (fm && Object.keys(fm).length === 0) delete data.field_mapping;

  return data;
}
