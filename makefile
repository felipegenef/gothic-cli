STAGE ?= default

# Commands
TEMPL_CMD = templ generate
GENERATE_SAM_TEMPLATE_CMD = go run .gothicCli/buildSamTemplate/main.go --stage $(STAGE)
CLEANUP_DEPLOY_FILES = go run .gothicCli/buildSamTemplate/cleanup/main.go
TAILWIND_CMD = ./tailwindcss -i src/css/app.css -o public/styles.css  --minify
SAM_BUILD_CMD = go run .gothicCli/sam/main.go --action build
SAM_DEPLOY_CMD = go run .gothicCli/sam/main.go --action deploy --stage $(STAGE)
SAM_DELETE_CMD = go run .gothicCli/sam/main.go --action delete --stage $(STAGE)
AWS_CP_CMD = go run .gothicCli/CdnAddOrRemoveAssets/main.go --stage $(STAGE) --action add
AWS_RM_CMD =  go run .gothicCli/CdnAddOrRemoveAssets/main.go --stage $(STAGE) --action delete
SERVE_APP_CMD = air
HOT_RELOAD_CMD = go run .gothicCli/HotReload/main.go
OPTIMIZE_CMD =  go run .gothicCli/imgOptimization/main.go
SETUP_OPTIMIZE_CMD = go run .gothicCli/imgOptimization/setup/main.go

# To deploy your app and public folder
deploy: 
	$(GENERATE_SAM_TEMPLATE_CMD)
	$(TEMPL_CMD)
	$(TAILWIND_CMD)
	$(SAM_BUILD_CMD)
	$(SAM_DEPLOY_CMD)
	$(SETUP_OPTIMIZE_CMD)
	${OPTIMIZE_CMD}
	$(AWS_CP_CMD)
	${CLEANUP_DEPLOY_FILES}

delete: 
	$(GENERATE_SAM_TEMPLATE_CMD)
	$(AWS_RM_CMD)
	$(SAM_DELETE_CMD)
	${CLEANUP_DEPLOY_FILES}

# Runs and watches your golang app
serve-app:
	$(SERVE_APP_CMD)	

# Starts your Application in dev mode watching Templates, golang files and CSS files
dev: 
	$(HOT_RELOAD_CMD)


# Compiles Templates to generate equivalent golang files 
templ:
	$(TEMPL_CMD)	

# Compiles Templates to generate equivalent golang files in watch mode
hot-reload-templ:
	$(TEMPL_CMD) --watch --proxy=http://localhost:8080

# Generate CSS based on classes located on templ files
css:
	$(TAILWIND_CMD)

# Generate CSS based on classes located on templ files in watch mode
hot-reload-css:
	$(TAILWIND_CMD) --watch 

# Generate CSS based on classes located on templ files in watch mode
optimize-images:
	$(SETUP_OPTIMIZE_CMD)
	$(OPTIMIZE_CMD)


