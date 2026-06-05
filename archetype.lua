local context = Context.new()

-- Identity
require("author").prompt(context)
require("org").prompt(context)

context:set("suffix_options", { "Service", "Orchestrator", "Adapter", "Router", "Gateway" })
context:set("suffix_default", "Service")
require("project").prompt(context)

context:set("repo_name", context:get("project-name"))
context:set("github_owner", context:get("org-solution-name"))

context:prompt_text("Go Module Path:", "module_path", {
    default = "github.com/" .. context:get("org-solution-name") .. "/" .. context:get("project-name"),
    help = "Go module path (e.g. github.com/p6m-acme/billing-service)",
})

-- Service configuration
require("ports").prompt(context)

-- Resources
context:prompt_select("Persistence:", "persistence", {
    "None", "PostgreSQL", "MySQL",
}, { default = "None" })

context:prompt_select("Cache:", "cache", {
    "None", "Redis",
}, { default = "None" })

context:prompt_select("Messaging:", "messaging", {
    "None", "Kafka", "Pulsar",
}, { default = "None" })

if context:get("messaging") ~= "None" then
    context:prompt_select("Messaging Access:", "messaging_access", {
        "produce", "consume",
    }, { default = "produce" })
else
    context:set("messaging_access", "produce")
end

context:prompt_multiselect("Object Storage:", "object_storage", {
    "S3", "Azure Blob",
}, { default = {} })

context:set("has_persistence", context:get("persistence") ~= "None")
context:set("has_cache",       context:get("cache")       ~= "None")
context:set("has_messaging",   context:get("messaging")   ~= "None")
context:set("has_s3",          context:contains("object_storage", "S3"))
context:set("has_azure_blob",  context:contains("object_storage", "Azure Blob"))

-- EditorConfig + gitignore (pre-seeded; library skips interactive prompt)
local editor_config = require("editor-config")
editor_config.prompt(context, {
    languages     = { "Go", "YAML", "Markdown" },
    gitattributes = true,
})

local gitignore = require("gitignore")
gitignore.prompt(context, {
    ignores = { "Go", "Claude", "IDEA", "VSCode", "macOS" },
})

-- SCM (opt-in: default = None)
local scm = require("scm")
scm.prompt(context)

if archetype.switches.is_enabled("debug-context") then
    log.info(archetype.description .. " Context:")
    output.print(format.yaml(context))
end

-- Render base workspace
directory.render("contents/base", context)

-- Resource libraries
local dest = { destination = context:get("project-name") }

if context:get("persistence") == "PostgreSQL" then
    require("golang-resource-postgresql").render(context, dest)
elseif context:get("persistence") == "MySQL" then
    require("golang-resource-mysql").render(context, dest)
end

if context:get("has_cache") then
    require("golang-resource-redis").render(context, dest)
end

if context:get("messaging") == "Kafka" then
    require("golang-resource-kafka").render(context, dest)
elseif context:get("messaging") == "Pulsar" then
    require("golang-resource-pulsar").render(context, dest)
end

if context:get("has_s3") then
    require("golang-resource-s3").render(context, dest)
end

if context:get("has_azure_blob") then
    require("golang-resource-azure-blob").render(context, dest)
end

-- CI workflows
local ci = require("golang-ci")
ci.render(context, dest)

-- Platform manifests
context:set("protocol", "REST")
local platform = require("platform-application-manifests")
platform.prompt(context)
platform.finalize(context, dest)

-- EditorConfig, gitignore, SCM finalize (side effects last)
editor_config.finalize(context, dest)
gitignore.finalize(context, dest)
scm.finalize(context)

-- Archive (zip / tarball switches for Ybor Studio)
require("archiver").finalize(context)

return context
