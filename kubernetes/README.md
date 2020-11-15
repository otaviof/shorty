Shorty on Kubernetes
--------------------

When using [`ko`][googleko], you can directly compile and deploy on Kubernetes with `ko apply`,
that is:

```sh
ko apply -f kubernetes/deployment.yaml
```

To produce a new relase image with `ko`, run:

```sh
ko resolve -P -f kubernetes/deployment.yaml > kubernetes/release.yaml
```

And later on you can deploy a pre-compiled image with:

```sh
kubectl apply -f kubernetes/release.yaml
```

[googleko]: https://github.com/google/ko