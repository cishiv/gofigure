# Basic CI/CD Pipeline for Local Dev

"CI/CD, but worse"

This is extremely specific to my setup for the moment.

My current development environment is using Docker and Minikube, with a custom .sh build script
to create my Go binaries and Docker images.

Even though the builds are fast and the scripts are pretty cool, it can be pretty manual and monotonous
for fast testing.

I wrote this little app that can run in a seperate tmux tab to `watch` the current directory and compute hashes on files. The `watch` time is set to 2 seconds which is insanely low, more reasonable would be around 1 - 3 minutes.

It uses [Bitcask](github.com/prologic/bitcask) as a key/value store to keep a registry of files and hashes. The files are the key since we use the absolute path. 

When a hash change is detected for a file, we run:

`./build build docker && kubectl delete --all deployments && kubectl delete --all services && kubectl apply -f deploy.yml` 

which is a set of commands specific to bootstrapping and enviroment for [Cortex](github.com/cishiv/Cortex) (currently a private repo) 

`./build build docker` is a custom build script for the Cortex project.

## Extensions

I will probably extend this and make it more configurable as needed. But it serves my purpose for the moment.
