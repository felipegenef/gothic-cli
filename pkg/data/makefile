STAGE ?= dev


# To deploy your app and public folder
deploy: 
	gothic-cli deploy --action deploy --stage $(STAGE)

delete: 
	gothic-cli deploy --action delete --stage $(STAGE)

# Starts your Application in dev mode watching Templates, golang files and CSS files
dev: 
	gothic-cli hot-reload


# Compiles Templates to generate equivalent golang files 
templ:
	gothic-cli build	


# Generate CSS based on classes located on templ files
css:
	./{{.TailWindFileName}} -i src/css/app.css -o public/styles.css  --minify

# Generate CSS based on classes located on templ files in watch mode
optimize-images:
	gothic-cli optimize-images


