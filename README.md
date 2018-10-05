# buildkite-signed-pipeline

This is a tool that adds some extra security guarantees around Buildkite's jobs. Buildkite [security best practices](https://buildkite.com/docs/agent/v3/securing) suggest using `--no-command-eval` which will only allow local scripts in a checked out repository to be run, preventing arbitrary commands being injected by an intermediary.

The downside of that approach is that it also comes with the recommendation of disabling plugins, or allow listing specifically what plugins and parameters are allowed. This tool is a collaboration between SEEK and Buildkite that attempts to bridge this gap and allow uploaded steps to be signed with a secret shared by all agents, so that plugins can run without any concerns of tampering by third-parties.

## Attack scenarios

For reference, this tool considers at least the following attack scenarios:

 1. A malicious user gains access to the Buildkite UI (buildkite.com), and updates pipeline settings (adds/modifies a command or plugin)
  - ✅ Commands cannot be signed without knowing the signing secret
 2. Buildkite is compromised and arbitrary jobs (commands) are sent to all known agents
  - ✅ Only pipelines from your repositories will be signed
  - [**Make sure your agents check `BUILDKITE_REPO` to ensure only known repositories are cloned**](https://buildkite.com/docs/agent/v3/securing#whitelisting)
 3. The command (`BUILDKITE_COMMAND`) for a job is changed by a man-in-the-middle between Buildkite.com and your agents
  - ✅ The job signature validation will fail as it will not match the command from the uploaded pipeline
 4. A plugin parameter is changed (e.g. `docker` `image`) by a man-in-the-middle to a poisoned Docker image
  - ✅ The job signature validation will fail as it will not match the plugin from the uploaded pipeline
 5. A malicious plugin is added to a known "allow listed command" by a man-in-the-middle
  - ✅ This tool requires that jobs with plugins are signed, regardless of the allowed command
 6. A malicious user gains access to your build agents
  - ❌ This tool will not help in this scenario
 7. The signing secret is leaked/stolen
  - ❌ With the right signing secret, any `command`/`plugins` combination can be signed (and thus trusted by your agents)
 8. A malicious user gains access to your allow-listed git repositories (e.g. on GitHub)
  - ❌ This tool will not help in this scenario

## Example

### Uploading a pipeline with signatures

Upload is a thin wrapper around [`buildkite-agent pipeline upload`](https://buildkite.com/docs/agent/v3/cli-pipeline#uploading-pipelines) that adds the required signatures. It behaves much like the command it wraps.

```bash
export SIGNED_PIPELINE_SECRET='my secret'

buildkite-signed-pipeline upload
```

### Verifying a pipeline signature

In a global `environment` hook, you can include the following to ensure that all jobs that are handed to an agent contain the correct signatures:

```bash
# Allow the upload command to be unsigned, as it typically comes from the Buildkite UI and not your agents
if [[ "${BUILDKITE_COMMAND}" == "buildkite-signed-pipeline upload" ]]; then
  echo "Allowing pipeline upload"
  exit 0
fi

export SIGNED_PIPELINE_SECRET='my secret'

if ! buildkite-signed-pipeline verify ; then
  echo "Step verification failed"
  exit 1
fi
```

This step will fail if the provided signatures aren't in the environment.

## Managing signing secrets

### Simple secret

Per the examples above, the secret for signing and verification can be provided via an environment variable or command line flag.

### AWS SM

This tool also has first-class support for [AWS Secrets Manager (AWS SM)](https://aws.amazon.com/secrets-manager/).
A secret id or ARN can be provided, the secret value will then be fetched from AWS SM to be used for signing and verification.

```bash
export SIGNED_PIPELINE_AWS_SM_SECRET_ID='arn:aws:secretsmanager:ap-southeast-2:12345:secret:my-signed-pipeline-secret-42a5qP'

buildkite-signed-pipeline upload
```

Future versions of the tool will add support for secret versioning.

## How it works

When the tool receives a pipeline for upload, it follows these steps:

* Iterates through each step of a JSON pipeline
* Extracts the `command` or `commands` block
* Trims whitespace on resulting command
* Calculates `HMAC(SHA256, command + canonicalised(BUILDKITE_PLUGINS), shared-secret)`
* Add `STEP_SIGNATURE={hash}` to the step `environment` block
* Pipes the modified JSON pipeline to `buildkite-agent pipeline upload`

When the tool is verifying a pipeline:

* Calculates `HMAC(SHA256, BUILDKITE_COMMAND + canonicalised(BUILDKITE_PLUGINS), shared-secret)`
* Compare result with `STEP_SIGNATURE`
* Fail if they don't match

## Development

This is using Golang's 1.11 modules.

```
export GO111MODULE=on
go run .
```
