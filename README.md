<img alt="background doc" src="Doc/Assets/background.jpeg" width="100%" height="280"/>

# <img alt="background doc" src="Doc/Assets/logo.jpeg" width="100" /> Gothic-cli

## The Tool

Gothic-cli is designed to help you create modern applications with Golang in a fast, simple, and scalable manner using the GOTTH stack (Golang, TailwindCSS, Templ, and HTMX).

This tool draws inspiration from Next.js features, particularly its ability to leverage serverless and edge environments to enhance user experience (UX) while also providing a positive developer experience (DX).

Gothic-cli generates boilerplates with some libraries pre-installed to assist you, but this does not limit your options. Feel free to choose the infrastructure, cloud services, caching solutions, and libraries that best suit your needs! After all, in the end it is just a Go app binary that can be deployed anywhere!

The currently selected libraries and default tools include:

- [AWS CloudFront](https://aws.amazon.com/cloudfront/) for CDN
- [AWS Lambda Container](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html) for server hosting
- [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html) for infrastructure as code (IaC) and deployment
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) to work alongside AWS SAM
- [CHI](https://go-chi.io) as the HTTP server framework
- [AIR](https://github.com/air-verse/air) for hot reloading
- [ShortId](github.com/teris-io/shortid) for stack bucket and lambda unique name creation
- [TailwindCSS](https://tailwindcss.com/) for styling
- [Chai Webp](github.com/chai2010/webp) for the image optimization tool on webp images
- [Nfnt resize](github.com/nfnt/resize) for the image optimization tool
- [HTMX](https://htmx.org/) for handling HTML events and rendering the DOM
- [Templ](https://templ.guide/) for creating HTML page templates

## Getting Started

To use tool it is as simple as installing and initialize your development.

### Install the latest version:

```bash
go install github.com/felipegenef/gothic-cli@latest
```

### Init the project in a folder that you wish to start

```bash
gothic-cli --init
```

### Start the application locally

```bash
make dev
```

## Simple to Deploy, Robust Infrastructure!

To ensure a smooth developer experience, we have designed the default infrastructure to be 100% serverless. This allows you to focus on web app development and business logic without worrying about infrastructure and scalability.

<img alt="background doc" src="Doc/Assets/Infrastructure.jpeg" width="100%"/>

### Deploy your Application to Test!

To make this automagically deployment happen with a robust infrastructure you will need to create an AWS account and install some tools:

#### Installing Tools to Deploy

1. Download [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).
2. Download [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).
3. Login with AWS CLI and add IAM credential to your CLI to work.

#### Deploy your first Gothic App

Now that we have all set, just add a new key `deploy` to your `gothic-config.json`. Dont worry for the values for now, leave them as blank. Just add the `deploy` config to your `gothic-config.json` and run:

##### Add Deploy config

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  },
  "deploy": {
    "serverMemory": 128,
    "serverTimeout": 30,
    "region": "us-east-1",
    "customDomain": false,
    "profile": "default",
    "stages": {
      "dev": {
        "hostedZoneId": null,
        "customDomain": null,
        "certificateArn": null,
        "env": {}
      }
    }
  }
}
```

##### Command to deploy

```bash
make deploy STAGE=dev
```

This will create the infrastructure we showed above. With CloudFront as a CDN, a Lambda Container as the server and an S3 bucket for the images on your public folder! A fully managed and robust application !

## Infrastructure

### CDN

We use AWS CloudFront for the CDN.

- All assets added to the `public` folder will be uploaded to an S3 bucket and linked to the CDN.
- All optimized images should be placed in the `optimize` folder. These images will be resized and stored in the S3 bucket as a separate folder. (For example, `optimized/logo.png` will generate `public/logo/blurred.png` and `public/logo/original.png`).

### Server

The server will be deployed as a Lambda Container, allowing it to scale rapidly according to your needs. It utilizes [Lambda Web Adapters](https://github.com/awslabs/aws-lambda-web-adapter) to handle incoming requests. This Lambda URL is connected to the same CloudFront CDN mentioned earlier.

### Caching

We also use [AWS CloudFront](https://aws.amazon.com/cloudfront/) for page caching. To cache your pages, add the "Cache-Control" header with your desired caching behavior. CloudFront will handle the caching of your pages and components at its edge locations.

### Infrastructure as Code (IaC)

With all recommended infrastructure hosted in AWS as serverless services, we will use AWS SAM to create and deploy our infrastructure.

## Amazing Next.js Like Features!

### SEO Optimized Image Loading (Next Image Component)

This feature optimizes SEO by implementing lazy loading for images, similar to the Image Component in Next.js. Initially, a lower-resolution version of the image is displayed. Once the page has loaded, the original image is fetched, giving the user the impression that the page loaded faster. The image will appear blurred at first, then transition smoothly to full resolution.

#### Add Image and Run Optimization Command

To use this feature, you will need two versions of your image: one with a lower resolution and another at the original resolution. Place your image in the _optimize_ folder and run (This command also run automatically on `make dev`):

```bash
make optimize-images
```

_Supported image formats include_:

- jpg
- jpeg
- webp
- png

This command executes the script located in _.gothicCli/imgOptimization/main.go_, which creates a folder in the public directory containing both the blurred and original images. By default, the blurred image is generated at 20% of the original resolution. You can adjust this by changing the variable in `gothic-config.json` file

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  }
}
```

Now that we have our optimized images on the `public` folder, we have to create the component template and route and add it to a page for the lazy loading effect similar to Next.js Image component.

#### Create the Image Component Template

_Example_

```go
templ OptimizedImage(isFirstLoad bool,imgName string, imgExtension string,alt string) {
        if isFirstLoad {
            <img alt={alt} class={"w-full h-full"} hx-trigger="load" hx-swap="outerHTML" hx-get={"/optimizedImage/"+imgName+"/"+imgExtension+"/"+alt} src={"/public/"+imgName+"/blurred."+imgExtension}/>
        }else{
            <img alt={alt} class={"w-full h-full"} src={"/public/"+imgName+"/original."+imgExtension}/>
        }

}
```

#### Add template component to a page

_Example_

```go
templ Index() {
		@layouts.PageLayout(){
			<div class="sm:w-[300px] sm:h-[300px] w-[200px] h-[200px]">
				@components.OptimizedImage(true,"imageExample","jpeg","image example alt text")
			</div>
		}

}
```

#### Create the Route to Render the component on the lazy load

_Example_

```go
	router.Get("/optimizedImage/{name}/{extension}/{alt}", func(w http.ResponseWriter, r *http.Request) {
		imgName := chi.URLParam(r, "name")
		imgExtension := chi.URLParam(r, "extension")
		imgAlt := chi.URLParam(r, "alt")
		handler.Render(r, w, components.OptimizedImage(false, imgName, imgExtension, imgAlt))
	})
```

### Public Static Pages CDN Caching

_Example:_

```go
router.Get("/cachedPageRoute", func(w http.ResponseWriter, r *http.Request) {
    currentTime := time.Now()
    // Max cache age for CloudFront is 31536000 = 1 year
    w.Header().Set("Cache-Control", "max-age=31536000")
    handler.Render(r, w, pages.Revalidate(currentTime))
})
```

### Incremental Static Regeneration in Public Pages (ISR)

_Example:_

```go
router.Get("/revalidateEvery10SecPage", func(w http.ResponseWriter, r *http.Request) {
    currentTime := time.Now()
    w.Header().Set("Cache-Control", "max-age=10, stale-while-revalidate=10, stale-if-error=10")
    handler.Render(r, w, pages.Revalidate(currentTime))
})
```

### Multi-Region (Edge Functions)

Currently, we have not implemented multi-region edge functions because AWS `@EdgeFunctions` do not support container images or Golang images. Please feel free to submit a pull request when this feature becomes available.

## And Much More!

### Secure your Environment Variables

You can get variables directly from AWS Parameter Store from the template using the _resolve:ssm_ as shown below. We recommend storing variables for each Stage. Please read more of how to do it in the next topic `Multi-Stage Deployments`

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  },
  "deploy": {
    "serverMemory": 128,
    "serverTimeout": 30,
    "region": "us-east-1",
    "customDomain": false,
    "profile": "default",
    "stages": {
      "dev": {
        "hostedZoneId": "{{resolve:ssm:/gothic-cli/dev/hostedZoneId}}",
        "customDomain": "dev.mycustomDomain.com",
        "certificateArn": "{{resolve:ssm:/gothic-cli/dev/certificateArn}}",
        "env": {}
      },
      "qa": {
        "hostedZoneId": "{{resolve:ssm:/gothic-cli/qa/hostedZoneId}}",
        "customDomain": "qa.mycustomDomain.com",
        "certificateArn": "{{resolve:ssm:/gothic-cli/qa/certificateArn}}",
        "env": {}
      },
      "prod": {
        "hostedZoneId": "{{resolve:ssm:/gothic-cli/prod/hostedZoneId}}",
        "customDomain": "mycustomDomain.com",
        "certificateArn": "{{resolve:ssm:/gothic-cli/prod/certificateArn}}",
        "env": {}
      }
    }
  }
}
```

### Multi-Stage Deployments

Add different stages and different variables for each stage on your `gothic-config.json`. After that you can deploy the same app In different stages with a simple command:

```bash
make deploy STAGE=yourCustomStage
```

### Deploy your App with your Custom Domain from AWS

Deploy your app with your custom domain is easy! You will just need your hostedZoneId from your AWS Route 53 hostedZone and the domain or subdomian of your choice. Once you have those two values add them to your `gothic-config.json` as shown below (we recommend storing your hosted zone id in Parameter Store as shown in the example as it is sensitive information):

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  },
  "deploy": {
    "serverMemory": 128,
    "serverTimeout": 30,
    "region": "us-east-1",
    "customDomain": false,
    "profile": "default",
    "stages": {
      "dev": {
        "hostedZoneId": "{{resolve:ssm:/gothic-cli/dev/hostedZoneId}}",
        "customDomain": "dev.mycustomDomain.com",
        "certificateArn": null,
        "env": {}
      }
    }
  }
}
```

#### Important Note!

If your app is in a different region than us-east-1 please also add an AWS ACM us-east-1 arn to `certificateArn` on your `gothic-config.json`. For more information, please go to our `Custom Region infrastructure` session.

#### Then deploy your app with custom domain

```bash
make deploy STAGE=dev

```

### Multiple AWS Account Profile Deployments

Sometimes you have more than one AWS account profile on your AWS CLI, to use a specific credential profile add the profile name to your
`gothic-config.json` as shown below:

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  },
  "deploy": {
    "serverMemory": 128,
    "serverTimeout": 30,
    "region": "us-east-1",
    "customDomain": false,
    "profile": "mycustomProfileName",
    "stages": {
      "dev": {
        "hostedZoneId": null,
        "customDomain": null,
        "certificateArn": null,
        "env": {}
      }
    }
  }
}
```

### Custom Region infrastructure

At present, deploying your functions in regions other than `us-east-1`, while also creating the ACM certificate in `us-east-1` within the same template, is not straightforward. For the Route 53 A record to work and for the CloudFormation CDN to have an alias domain, the ACM certificate must be created in the us-east-1 region. If you want to create your infrastructure in another region, such as `eu-central-1` (Central Europe), you will need to _manually create your ACM certificate in the AWS console_ and _reference it in `gothic-config.json` as an ARN_ value (we recommend storing it in Parameter Store).

Although this limitation may not be an issue for most use-cases since you can optimize your websites with CDN caching for pages and components, some scenarios may require the infrastructure to be deployed in a specific region. Here’s an example of how to do it. First, you change your region in your `gothic-config.json`, then create your certificate for your domain in the `us-east-1` region and validate it in your Route 53 hosted zone. For last, add your ACM certificate ARN to SSM and include this in your `gothic-config.json`.

#### Change your region on `gothic-config.json`

```json
{
  "projectName": "gothic-example",
  "goModuleName": "github.com/felipegenef/gothic-cli",
  "optimizeImages": {
    "lowResolutionRate": 20
  },
  "deploy": {
    "serverMemory": 128,
    "serverTimeout": 30,
    "region": "eu-central-1",
    "customDomain": false,
    "profile": "default",
    "stages": {
      "dev": {
        "hostedZoneId": "{{resolve:ssm:/gothic-cli/dev/hostedZoneId}}",
        "customDomain": "dev.mycustomDomain.com",
        "certificateArn": "{{resolve:ssm:/gothic-cli/dev/certificateArn}}",
        "env": {}
      }
    }
  }
}
```

#### deploy your application

```bash
make deploy STAGE=dev
```

## TODOs

- [x] SEO Optimized Image Load
- [x] Public Static Pages CDN Caching
- [x] Incremental Static Regeneration in Public Pages (ISR)
- [x] Hot Reload locally
- [x] Fetch environment variables from Parameter Store
- [x] Multi-Stage Deployments
- [x] Custom Domain
- [x] Custom Region infrastructure
- [x] CLI creating boilerplates for basic component, pages and api routes
- [x] Delete or set a limit for old ECR images
- [x] Simple Config File
- [x] Config create SAM templates based on json and easy deploy templating
- [ ] Multi-Region (Edge Functions)
- [ ] Website Docs
- [ ] Integrated WAF to Cloudfront Distribution
