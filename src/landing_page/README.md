# Landing Page

## Local Build

You have two options for building the landing page:

- Use Bazel
- Use Jekyll directly.

With both options, you will be able to go to `http://localhost:4000` to view the landing page.

### Bazel

- `cd $GRCHIVE`
- `bazel run //src/landing_page:site`

If you wish to build the production website (i.e. with Google Analytics):

- `bazel build -c opt //src/landing_page:site`

Note that the `-c opt` option will have no effect on running the site locally.

### Jekyll

- `cd $GRCHIVE/src/landing_page`
- `jekyll serve`

If you wish to run the production website (i.e. with Google Analytics):

- `JEKYLL_ENV=production jekyll serve`

### Nginx

- `bazel run //src/landing_page:latest`
- `docker run --env DISABLE_CERTBOT=1 -p 80:80 bazel/src/landing_page:latest`
