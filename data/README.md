<img alt="background doc" src="Doc/Assets/background.jpeg" width="100%" height="280"/>

# <img alt="background doc" src="Doc/Assets/logo.jpeg" width="100" /> Gothic-cli

## The Tool

Gothic-cli is designed to help you create modern applications with Golang in a fast, simple, and scalable manner using the GOTTH stack (Golang, TailwindCSS, Templ, and HTMX).

This tool draws inspiration from Next.js features, particularly its ability to leverage serverless and edge environments to enhance user experience (UX) while also providing a positive developer experience (DX).

Gothic-cli generates boilerplates with some libraries pre-installed to assist you, but this does not limit your options. Feel free to choose the infrastructure, cloud services, caching solutions, and libraries that best suit your needs! After all, in the end it is just a Go app binary that can be deployed anywhere!

[![See our docs](https://img.shields.io/badge/See_our_docs-ec4899?style=for-the-badge)](https://gothicframework.com)

### Default Tools and Libraries

The following libraries and tools are included by default:

- [AWS CloudFront](https://aws.amazon.com/cloudfront/) for CDN
- [AWS Lambda Container](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html) for server hosting
- [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html) for infrastructure as code (IaC) and deployment
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) for use with AWS SAM
- [CHI](https://go-chi.io) as the HTTP server framework
- [AIR](https://github.com/air-verse/air) for hot reloading
- [ShortId](github.com/teris-io/shortid) for creating unique names for stacks and Lambda functions
- [TailwindCSS](https://tailwindcss.com/) for styling
- [X image](golang.org/x/image) for WebP image optimization
- [Nfnt Resize](github.com/nfnt/resize) for image resizing and optimization
- [HTMX](https://htmx.org/) for handling HTML events and DOM updates
- [Templ](https://templ.guide/) for HTML page templating

## Simple to Deploy, Robust Infrastructure!

To ensure a smooth developer experience, we have designed the default infrastructure to be 100% serverless. This allows you to focus on web app development and business logic without worrying about infrastructure and scalability.

<img alt="background doc" src="Doc/Assets/Infrastructure.jpeg" width="100%"/>

## Amazing Next.js Like Features!

### SEO-Optimized Image Loading

Gothic-cli includes a feature that improves SEO by implementing lazy-loading for images, similar to the Next.js Image component. Initially, a lower-resolution image is shown, which is then replaced by the original image after the page loads. This gives the appearance of faster loading times and smooth image transitions.

### Static Page CDN Caching

Like Next.js, static pages created with Gothic-cli can be cached on CloudFront Edge locations for fast delivery, with a time to live (TTL) up to 1 year!

### Incremental Static Regeneration (ISR)

Gothic-cli supports Incremental Static Regeneration (ISR) for public pages. You can specify the revalidation time, which can be set up to 1 year.

### Custom 404 Pages

You can create a custom 404 page for situations where a user enters an incorrect URL or when a page is no longer available, enhancing the user experience.

### Link Prefetching

Similar to Next.js’s Link component, Gothic-cli enables link prefetching on mouseover events. This allows pages to be preloaded in the background, so when the user clicks the link, the page loads instantly, improving navigation speed.

## And Much More!

### Secure Environment Variables

Gothic-cli allows you to securely retrieve environment variables directly from AWS Parameter Store, ensuring sensitive information is not exposed in your code.

### Multi-Stage Deployments

You can define multiple stages and variables for each stage in your `gothic-config.json` file. This makes it easy to deploy the same app to different environments with a single command.

### Deploy with a Custom Domain from AWS

Deploying your app with a custom domain is simple. Just set the `customDomain` flag to `true` in your `gothic-config.json` file. You’ll need the `hostedZoneId` from AWS Route 53 and the domain (or subdomain) of your choice.

#### Important Note

If your app is hosted in a region other than `us-east-1`, you’ll need to add an AWS ACM certificate ARN from `us-east-1` to your `gothic-config.json` file. For more information, refer to the "Custom Region Infrastructure" section below.

### Multiple AWS Account Profile Deployments

If you have multiple AWS account profiles set up in your AWS CLI, you can specify which profile to use by adding the profile name to your `gothic-config.json`.

### Custom Region Infrastructure

At present, deploying your functions in regions other than us-east-1, while also creating the ACM certificate in us-east-1 within the same template, is not straightforward. For the Route 53 A record to work and for the CloudFormation CDN to have an alias domain, the ACM certificate must be created in the us-east-1 region. If you want to create your infrastructure in another region, such as eu-central-1 (Central Europe), you will need to manually create your ACM certificate in the AWS console and reference it in gothic-config.json as an ARN value (we recommend storing it in Parameter Store).
