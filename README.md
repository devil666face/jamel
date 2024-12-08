```
podman:yourrepo/yourimage:tag
docker:yourrepo/yourimage:tag
docker-archive:path/to/yourimage.tar
oci-archive:path/to/yourimage.tar
oci-dir:path/to/yourimage
singularity:path/to/yourimage.sif
dir:path/to/yourproject
file:path/to/yourfile
sbom:path/to/syft.json
registry:yourrepo/yourimage:tag
```

docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock ubuntu:latest bash
