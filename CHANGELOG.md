# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Normally I'd adhere to [Semantic Versioning](https://semver.org/spec/v2.0.0.html), 
but since this is essentially an incremental lab notebook I'll use a simpler date 
based notation instead.

## [2022.07.22.1]

## Changed
* The [service controller](controllers/service_controller.go) now uses a `Patch` call to
update the related Service object, instead of an `Update` call, which might overwrite other
aspects of the service (if, for example, something else updated it too).

[2022.07.22.1]: https://github.com/yardbirdsax/k8s-controller/compare/2022.07.21.1...2022.07.22.1
