STAGE ?= default
S3_BUCKET ?= gothic-example-public-bucket-$(STAGE)

# Commands
TEMPL_CMD = templ generate
TAILWIND_CMD = ./tailwindcss -i src/css/app.css -o public/styles.css  --minify
SAM_BUILD_CMD = sam build --config-file .gothicCli/deploy/samconfig.toml
SAM_DEPLOY_CMD = sam deploy --stack-name gothic-example-${STAGE} --parameter-overrides Stage=$(STAGE) --config-file .gothicCli/deploy/samconfig.toml
SAM_DELETE_CMD = sam delete --stack-name gothic-example-${STAGE} --config-file .gothicCli/deploy/samconfig.toml
AWS_CP_CMD = aws s3 cp public s3://$(S3_BUCKET)/public --recursive
AWS_RM_CMD =  aws s3 rm s3://$(S3_BUCKET)/public --recursive
SERVE_APP_CMD = air
HOT_RELOAD_CMD = go run .gothicCli/HotReload/main.go
OPTIMIZE_CMD = go run .gothicCli/imgOptimization/main.go

# To deploy your app and public folder
deploy: 
	$(TEMPL_CMD)
	$(TAILWIND_CMD)
	$(SAM_BUILD_CMD)
	$(SAM_DEPLOY_CMD)
	${OPTIMIZE_CMD}
	$(AWS_CP_CMD)

delete: 
	$(AWS_RM_CMD)
	$(SAM_DELETE_CMD)

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
	$(OPTIMIZE_CMD)


