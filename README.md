# mutating-webhook

## Table of Contents

1. [To do](#to-do)

## To do

* Validate certificate logic (do we have a CA private-key, if not, don't generate certificate, etc.)
* Add 5-10 minutes at random to generated cert to prevent all instance's certs from expiring at the same time.
* Determine certificate expiration and automatically stop the service when certificate expires. (optional?)
* dynamic reloading of configuration
    * Hash configmap, re-read/hash file every so often, if changes stop or re-run init?
* Add unit tests
* Add documentation