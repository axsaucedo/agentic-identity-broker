package docker

import rego.v1

default allow := false

allow if input.local

# --- IMAGE RULES ---

# Internal registry: require digest + provenance attestation
approved_registries := [
    "container-registry.zalando.net",
    "registry.opensource.zalan.do"
]

allow if {
    input.image.host in approved_registries
    #input.image.isCanonical
    #input.image.hasProvenance
}

# --- HTTP DOWNLOAD RULES ---

# Only allow downloads from approved internal domains with HTTPS
approved_download_hosts := []

allow if {
    input.http.schema == "https"
    input.http.host in approved_download_hosts
}

# --- GIT RULES ---

# Internal repos: allow freely
internal_git_prefixes := {
  "https://github.com/zalando-infosec/",
  "https://github.com/zalando-build/"
}

is_internal_git_repo if {
  some p in internal_git_prefixes
  startswith(input.git.remote, p)
}

allow if {
    input.git
    is_internal_git_repo
}

# --- ERROR MESSAGES ---

# Priority 1: The Registry itself is forbidden
deny_msg contains msg if {
    not input.local
    not input.image.host in approved_registries
    msg := sprintf("FORBIDDEN REGISTRY: '%s' is not an approved source.", [input.image.host])
}

# Priority 2: Registry is okay, but security standards aren't met
deny_msg contains msg if {
    input.image.host in approved_registries
    not input.image.isCanonical
    msg := "MISSING DIGEST: Images from Zalando registries must use a @sha256 digest, not a tag (like :latest)."
}

# Priority 3: Registry & Digest are okay, but provenance is missing
deny_msg contains msg if {
    input.image.host in approved_registries
    input.image.isCanonical
    object.get(input.image, "hasProvenance", false) == false
    msg := "MISSING PROVENANCE: This image was not built with SLSA attestations."
}

# CRITICAL: This mapping tells BuildKit WHAT to print
decision := {
    "allow": allow,
    "deny_msg": deny_msg
}