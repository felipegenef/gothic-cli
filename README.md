# Gothic-cli

A Golang tool for building modern web applications inspired by Next.js, utilizing the GOTTH stack and AWS SAM.

## The Tool

Gothic-cli is designed to help you create modern applications with Golang in a fast, simple, and scalable manner using the GOTTH stack (Golang, TailwindCSS, Templ, and HTMX).

This tool draws inspiration from Next.js features, particularly its ability to leverage serverless and edge environments to enhance user experience (UX) while also providing a positive developer experience (DX).

Gothic-cli generates boilerplates with some libraries pre-installed to assist you, but this does not limit your options. Feel free to choose the infrastructure, cloud services, caching solutions, and libraries that best suit your needs!

The currently selected libraries and default tools include:

- [AWS CloudFront](https://aws.amazon.com/cloudfront/) for CDN
- [AWS Lambda Container](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html) for server hosting
- [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html) for infrastructure as code (IaC) and deployment
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) to work alongside AWS SAM
- [CHI](https://go-chi.io) as the HTTP server framework
- [AIR](https://github.com/air-verse/air) for hot reloading
- [TailwindCSS](https://tailwindcss.com/) for styling
- [HTMX](https://htmx.org/) for handling HTML events and rendering the DOM
- [Templ](https://templ.guide/) for creating HTML page templates

## Installing

1. Download [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).
2. Download [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).
3. Download [Golang](https://go.dev/doc/install).
4. Install the CLI using the following command:

```bash
go install github.com/felipegenef/gothic-cli@latest
```

5. Create your template

```bash
gothic-cli --init
```

6. Run your server locally

```bash
make hot-reload
```

7. deploy your server

```bash
make deploy
```

## Infrastructure

To ensure a smooth developer experience, we have designed the default infrastructure to be 100% serverless. This allows you to focus on web app development and business logic without worrying about infrastructure and scalability.

### CDN

We use AWS CloudFront for the CDN.

- All assets added to the `public` folder will be uploaded to an S3 bucket and linked to the CDN.
- All optimized images should be placed in the `optimize` folder. These images will be resized and stored in the S3 bucket as a separate folder. (For example, `optimized/logo.png` will generate `public/logo/blurred.png` and `public/logo/original.png`).

Supported image formats include:

- jpg
- jpeg
- webp
- png

### Server

The server will be deployed as a Lambda Container, allowing it to scale rapidly according to your needs. It utilizes [Lambda Web Adapters](https://github.com/awslabs/aws-lambda-web-adapter) to handle incoming requests. This Lambda URL is connected to the same CloudFront CDN mentioned earlier.

### Caching

We also use [AWS CloudFront](https://aws.amazon.com/cloudfront/) for page caching. To cache your pages, add the "Cache-Control" header with your desired caching behavior. CloudFront will handle the caching of your pages and components at its edge locations.

### Infrastructure as Code (IaC)

With all recommended infrastructure hosted in AWS as serverless services, we will use AWS SAM to create and deploy our infrastructure.

## Features

### Hot Reload

To enable hot reloading of your application and see changes made to your templates and styling, run:

```bash
make hot-reload
```

This command will run AIR and Templ proxy simultaneously, allowing you to hot reload your application. For customization options, please refer to the _.air.toml_ file and the _CLI/HotReload/main.go_ file.

### SEO Optimized Image Loading

This feature optimizes SEO by implementing lazy loading for images, similar to the Image Component in Next.js. Initially, a lower-resolution version of the image is displayed. Once the page has loaded, the original image is fetched, giving the user the impression that the page loaded faster. The image will appear blurred at first, then transition smoothly to full resolution.

#### Add Image and Run Optimization Command

To use this feature, you will need two versions of your image: one with a lower resolution and another at the original resolution. Place your image in the _optimize_ folder and run:

```bash
make optimize-images
```

This command executes the script located in _CLI/imgOptimization/main.go_, which creates a folder in the public directory containing both the blurred and original images. By default, the blurred image is generated at 20% of the original resolution. You can adjust this by changing the variable in the script:

```go
// Calculate new dimensions for blurred image (20% of original)
newWidth := originalWidth * 20 / 100
newHeight := originalHeight * 20 / 100
```

#### Create the Image Component

_Example_

```go
templ OptimizedImage(isFirstLoad bool, imgName string, imgExtension string) {
    if isFirstLoad {
        <img class={"min-w-[500px] min-h-[500px]"} hx-trigger="load" hx-swap="outerHTML" hx-get={"/optimizedImage/"+imgName+"/"+imgExtension} src={"/public/"+imgName+"/blurred."+imgExtension}/>
    } else {
        <img class={"min-w-[500px] min-h-[500px]"} src={"/public/"+imgName+"/original."+imgExtension}/>
    }
}
```

#### Create the Page and Add the Image Component

_Example_

```go
templ Index() {
    @layouts.PageLayout() {
        @components.OptimizedImage(true, "imageExample", "webp")
    }
}
```

#### Create the Route to Render the Original Resolution Image

_Example_

```go
router.Get("/optimizedImage/{name}/{extension}", func(w http.ResponseWriter, r *http.Request) {
    imgName := chi.URLParam(r, "name")
    imgExtension := chi.URLParam(r, "extension")
    handler.Render(r, w, components.OptimizedImage(false, imgName, imgExtension))
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

### Multi-Region (Edge) Functions

Currently, we have not implemented multi-region functions, as AWS Edge Functions do not support container images or Golang images. Please feel free to submit a pull request when this feature becomes available.

## TODOs

- [x] SEO Optimized Image Load
- [x] Public Static Pages CDN Caching
- [x] Incremental Static Regeneration in Public Pages (ISR)
- [x] Hot Reload locally
- [x] Fetch environment variables from Parameter Store
- [ ] Custom Domain
- [ ] Delete or set a limit for old ECR images
- [ ] Multi-Region (Edge)
