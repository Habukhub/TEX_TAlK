function extract_env(tag, timestamp, record)
    local env = record["environment"] or "local"
    record["_logstash_prefix"] = "logs-tex-" .. env
    return 2, timestamp, record
end
