--- The p6m platform standards, held against this archetype (see prova-p6m-standards/docs/
--- standards.md). One suite, parameterized by the SAME answers the archetype renders from — the
--- expectations come from the `p6m` oracle, never from this file, so every language's archetype is
--- held to the identical bar. Two name shapes on purpose: casing bugs only show on the second.
---
--- Container-first (S8): the SUT is the archetype's own .platform/docker/prd image on a topology
--- network — docker is the only requirement, no Go toolchain on the host.

local p6m = require("p6m")
local postgres = require("postgres")
local mysql = require("mysql")

local SRC = "."

-- Two variants cover both axes cheaply: name shape (single vs multi-word — casing bugs only show
-- on the second) paired with persistence backend (PostgreSQL vs MySQL). Rows land in the
-- entity-named table `{prefix_name}s`, so the count SQL is derived from the same identity oracle.
local VARIANTS = {
	{
		prefix = "Customer",
		persistence = "PostgreSQL",
		db = postgres,
		count_by_name = function(id)
			return "SELECT count(*) FROM " .. id.prefix_name .. "s WHERE display_name = $1"
		end,
	},
	{
		prefix = "User Details",
		persistence = "MySQL",
		db = mysql,
		count_by_name = function(id)
			return "SELECT count(*) FROM `" .. id.prefix_name .. "s` WHERE display_name = ?"
		end,
	},
}

for _, n in ipairs(VARIANTS) do
	local id = p6m.identity{ prefix = n.prefix }
	local label = "standards[" .. id.project_name .. "/" .. n.persistence .. "]"
	local count_by_name = n.count_by_name(id)

	local project = prova.fixture(label .. ":project", Scope.File, function(ctx)
		return archetect.render{
			source = SRC,
			answers = {
				author_name = "Test Author",
				author_email = "test@example.com",
				org_name = "acme",
				solution_name = "platform",
				prefix_name = id.answers.prefix_name,
				suffix_name = id.answers.suffix_name,
				image_registry = "ghcr.io/acme",
				persistence = n.persistence,
			},
			destination = ctx:tempdir(),
			defaults = true,
		}
	end)

	local sut = prova.topology(label .. ":sut", function(ctx)
		local root = ctx:use(project):dir(id.project_name)
		return p6m.sut(ctx, { root = root.path, id = id, transport = "rest", db = n.db })
	end)

	prova.group(label, { requires = { "docker" }, tags = { "standards" } }, function(g)
		p6m.standards.api(g, sut, {
			persisted = function(t, name, count)
				local svc = t:use(sut)
				t:expect(
					svc.db.client:query_value(count_by_name, { name }),
					"rows in the database"
				):equals(count)
			end,
		})
		p6m.standards.runtime(g, sut)
	end)
end

-- CI parity (S10): the rendered project's own Build workflow path — the community-action inline
-- steps (go build/test) on a fresh clone, in the toolchain image. The Dockerfile and CI are two
-- independent build paths; S10 holds the second. One hollow render suffices: resource variants
-- change dependencies, not the command path.
local ci_project = prova.fixture("standards[golang-rest/None]:project", Scope.File, function(ctx)
	return archetect.render{
		source = SRC,
		answers = {
			author_name = "Test Author",
			author_email = "test@example.com",
			org_name = "acme",
			solution_name = "platform",
			prefix_name = "Example",
			suffix_name = "Service",
			image_registry = "ghcr.io/acme",
			persistence = "None",
		},
		destination = ctx:tempdir(),
		defaults = true,
	}
end)

prova.group("golang-rest[None]:ci", { requires = { "docker" }, tags = { "standards" } }, function(g)
	p6m.standards.ci_parity(g, ci_project, {
		stack = "golang",
		project_dir = "example-service",
		name = "golang-rest",
	})
end)

-- E7's released-tag bar, as the `p6m-pin` reminder: DUE while the manifest pins `dev` (the
-- YP6M-3372 staging window), silent again once the pin returns to a released tag. Heed it
-- (`prova --heed=p6m-pin`) when the window closes.
p6m.pin_reminder()
