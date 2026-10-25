# SOGo Docker Container

[![Release](https://img.shields.io/github/v/release/sonroyaalmerol/docker-sogo)](https://github.com/sonroyaalmerol/docker-sogo/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/sonroyaalmerol/docker-sogo)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo)
[![Docker Stars](https://img.shields.io/docker/stars/sonroyaalmerol/docker-sogo)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo)
[![Image Size](https://img.shields.io/docker/image-size/sonroyaalmerol/docker-sogo/latest)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo/tags)
[![Development](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/develop.yml/badge.svg)](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/develop.yml)
[![Release](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml/badge.svg)](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/sogo&style=flat-square)](https://artifacthub.io/packages/search?repo=sogo)
[![OCI security profiles](https://img.shields.io/badge/oci%3A%2F%2F-sogo-blue?logo=kubernetes&logoColor=white&style=flat-square)](https://github.com/sonroyaalmerol/docker-sogo/packages)

## What is SOGo?

SOGo is a fully supported and trusted groupware server with a focus on scalability and open standards. SOGo is released under the GNU GPL/LGPL v2 and above.

SOGo provides a rich AJAX-based Web interface and supports multiple native clients through the use of standard protocols such as CalDAV, CardDAV and GroupDAV, as well as Microsoft ActiveSync.

SOGo is the missing component of your infrastructure; it sits in the middle of your servers to offer your users a uniform and complete interface to access their information. It has been deployed in production environments where thousands of users are involved.

## Container Features

> [!IMPORTANT]
> Since July 2016, SOGo [limits their stable production builds](https://www.sogo.nu/news/2016/sogo-package-repositories.html) to those who pay for their support subscription.
>
> This container builds the stable source code and should be similar, if not the same, as the official stable builds. However, if you can afford [paying](https://www.sogo.nu/commercial.html), you should consider supporting SOGo.

This container is mainly built with Kubernetes in mind. As such, one of the main unique features of this container is its ability to use YAML configuration files.

- **Automated Builds**: SOGo **stable** releases are checked every day for updates and will trigger the Docker image build automatically.
- **Follows the official SOGo version tagging**: Docker image tags follows the version tagging of SOGo releases starting from version `5.10.0`. Backporting to different versions will be considered.
- **YAML Configuration Support**: Users can define SOGo configurations using YAML files, allowing for easier management and version control. The included script automatically merges and converts YAML configurations into the required OpenStep plist format, simplifying the setup process.

## Supported tags

To make things simpler, the container will mainly follow SOGo's versioning with the addition of a `revision` tag which will serve as an additional incremental versioning system for each container-specific modifications I push through a certain SOGo version. This way, features that I implement for the container can and will be backported to previous SOGo versions while having automated builds for new SOGo releases at the same time. For stability and pinning, you will want to use the tag with the container revision.

Unsuffixed tags are **vanilla** (upstream SOGo sources, no modifications). Tags suffixed with `-patched` carry the extra patches described in [Patches](#patches).

- `latest` (will always follow the latest container revision of the latest SOGo version, vanilla)
- `${SOGo-version}` (e.g. `5.12.12`, vanilla)
- `${SOGo-version}-${Container-Revision}` (e.g. `5.12.12-1`, vanilla)
- `${SOGo-version}-patched` (e.g. `5.12.12-patched`)
- `${SOGo-version}-${Container-Revision}-patched` (e.g. `5.12.12-1-patched`)

## Why did I build this container?

- Mainly due to SOGo being used by the company I work for. As we transition to using Kubernetes for our services, we needed to containerize most of our legacy services, including SOGo.
- We also needed a clear way to downgrade to a specific version as much as possible which proved to be difficult to do with currently available SOGo containers being built with nightly Debian packages.
- Using OpenStep plist format for configuration was difficult to maintain since we preferred to use multiple files for different sections of the configuration. With our database secrets in Kubernetes being dynamically generated, having one config file for database credentials separated from the rest of the configurations was essential.

## Usage

### Running the Docker Container

Assuming you have your YAML configuration files ready, you can run the Docker container using the following command:

```bash
docker run -v /path/to/sogo/configs:/etc/sogo/sogo.conf.d/ sonroyaalmerol/docker-sogo
```

Replace `/path/to/sogo/configs` with the directory containing your YAML configuration files.

The NGINX process inside the container will host the web server on port 80 by default. See the [default NGINX config](https://github.com/sonroyaalmerol/docker-sogo/blob/main/default-configs/nginx.conf) for more info.

### Running in Kubernetes with Helm

> [!IMPORTANT]
> This chart is still in its beta stage. Do not use in production.

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repository as follows:

```console
helm repo add sogo https://helm.snry.xyz/docker-sogo/
```

Running `helm search repo sogo` should now display the chart and it's versions

To install the helm chart, use

```console
helm install sogo sogo/sogo --create-namespace --namespace sogo
```

#### Values

You can find the `values.yaml` summary in [the charts directory](https://github.com/sonroyaalmerol/docker-sogo/blob/main/charts/sogo/values.yaml). The SOGo configurations in YAML can be placed in the `values.yaml` file in `.sogo.configs` as set in the default values. See how to convert your current SOGo configs to YAML with examples below.

### Configuration Examples

> [!IMPORTANT]
> YAML configuration files are NOT officially supported by SOGo.
>
> If you have any problems with using YAML configuration files, please open an issue in this repository.

For example, we have these two YAML config files mounted to `/etc/sogo/sogo.conf.d/`.

```yaml filename="database.yaml"
# Database configuration (mysql://, postgresql:// or oracle://)
SOGoProfileURL: "postgresql://sogo:sogo@localhost:5432/sogo/sogo_user_profile"
OCSFolderInfoURL: "postgresql://sogo:sogo@localhost:5432/sogo/sogo_folder_info"
OCSSessionsFolderURL: "postgresql://sogo:sogo@localhost:5432/sogo/sogo_sessions_folder"
```

```yaml filename="mail.yaml"
# Mail
SOGoDraftsFolderName: Drafts
SOGoSentFolderName: Sent
SOGoTrashFolderName: Trash
SOGoJunkFolderName: Junk
SOGoIMAPServer: "localhost"
SOGoSieveServer: "sieve://127.0.0.1:4190"
SOGoSMTPServer: "smtp://127.0.0.1"
SOGoMailDomain: "acme.com"
SOGoMailingMechanism: smtp
SOGoForceExternalLoginWithEmail: false
SOGoMailSpoolPath: "/var/spool/sogo"
NGImap4AuthMechanism: plain
NGImap4ConnectionStringSeparator: "/"
```

When the container initializes, the config-generator merges both YAML files and generates the final `/etc/sogo/sogo.conf` file.

```conf filename="sogo.conf"
{
  /* Database configuration (mysql://, postgresql:// or oracle://) */
  SOGoProfileURL = "postgresql://sogo:sogo@localhost:5432/sogo/sogo_user_profile";
  OCSFolderInfoURL = "postgresql://sogo:sogo@localhost:5432/sogo/sogo_folder_info";
  OCSSessionsFolderURL = "postgresql://sogo:sogo@localhost:5432/sogo/sogo_sessions_folder";

  /* Mail */
  SOGoDraftsFolderName = Drafts;
  SOGoSentFolderName = Sent;
  SOGoTrashFolderName = Trash;
  SOGoJunkFolderName = Junk;
  SOGoIMAPServer = "localhost";
  SOGoSieveServer = "sieve://127.0.0.1:4190";
  SOGoSMTPServer = "smtp://127.0.0.1";
  SOGoMailDomain = acme.com;
  SOGoMailingMechanism = smtp;
  SOGoForceExternalLoginWithEmail = NO;
  SOGoMailSpoolPath = /var/spool/sogo;
  NGImap4AuthMechanism = "plain";
  NGImap4ConnectionStringSeparator = "/";
}
```

The config parameters are exactly the same, just in a different format. Arrays in YAML will be converted to arrays in OpenStep plist. Maps in YAML will be converted to maps in OpenStep plist. Booleans in YAML will be converted to either `YES` or `NO` appropriately.

This YAML file below is equivalent to

```yaml
SOGoUserSources:
  - type: ldap
    CNFieldName: cn
    UIDFieldName: uid
    IDFieldName: uid
    bindFields: [uid, mail]
    baseDN: "ou=users,dc=acme,dc=com"
    bindDN: "uid=sogo,ou=users,dc=acme,dc=com"
    bindPassword: qwerty
    canAuthenticate: true
    displayName: "Shared Addresses"
    hostname: "ldap://127.0.0.1:389"
    id: public
    isAddressBook: true
```

this OpenStep plist:

```conf
{
  SOGoUserSources = (
    {
      type = ldap;
      CNFieldName = cn;
      UIDFieldName = uid;
      IDFieldName = uid;
      bindFields = (uid, mail);
      baseDN = "ou=users,dc=acme,dc=com";
      bindDN = "uid=sogo,ou=users,dc=acme,dc=com";
      bindPassword = qwerty;
      canAuthenticate = YES;
      displayName = "Shared Addresses";
      hostname = "ldap://127.0.0.1:389";
      id = public;
      isAddressBook = YES;
    }
  );
}
```

If you want to use the usual OpenStep plist config instead, just mount your existing `sogo.conf` to `/etc/sogo/sogo.conf` and make sure to leave `/etc/sogo/sogo.conf.d/` empty.

## Running multiple replicas

SOGo keeps sessions, login caches and group lookups in memcached, so every replica must point at the same memcached server. The chart fails to render if you ask for more than one replica without one, rather than letting pods drift apart silently:

```
running more than one replica requires memcached: either set memcached.enabled=true,
or point sogo.configs.SOGoMemcachedHost at an external server
```

### Scheduled jobs

Every replica mounts the same crontab, so at `replicaCount: 3` each entry fires three times a minute: three copies of every calendar alarm mail, three concurrent `sogo-backup.sh` runs onto one PVC. Each cron line is therefore wrapped in a memcached lock:

```
* * * * * sogo /usr/sbin/sogo-tool-plus -mode lock -key ealarms -ttl 50 -- /usr/sbin/sogo-ealarms-notify
```

The first replica to `ADD` the key runs the job, the rest log a skip and exit 0. The lock is released as soon as the job finishes, and the TTL bounds it so a pod that is killed mid-run cannot wedge the schedule. Missing or unreachable memcached fails open and runs the job, so a cache outage never silently stops alarm mail on a single-replica install. Set `sogo.cronSingleton.enabled: false` to go back to unwrapped commands.

### Drafts and attachments

In-progress drafts and their uploaded attachments are written to local disk under `SOGoMailSpoolPath`, not to IMAP or the database. A compose started on one pod is therefore incomplete on another. Pick one of:

- `service.sessionAffinity: ClientIP` to pin a client to one replica. Note this is applied by kube-proxy, so it does nothing when an ingress controller load-balances to pods itself; use that controller's sticky-session annotation instead.
- `sogo.spool.enabled: true` with a `ReadWriteMany` volume, which shares the spool across pods and needs no affinity. The chart refuses `ReadWriteOnce` here when `replicaCount > 1`.

```yaml
sogo:
  spool:
    enabled: true
    path: /var/spool/sogo
    persistence:
      accessMode: ReadWriteMany
      size: 8Gi
```

## Patches

The following patches are applied to the SOGo source at build time. They are located in `patches/` and applied in sorted order. Prebuilt images carrying them are published under the `-patched` tags (see [Supported tags](#supported-tags)); unsuffixed tags are vanilla.

### Skipping Patches

Set the `SKIP_PATCHES` build argument to `true` to build an unmodified SOGo:

```bash
docker build --build-arg SKIP_PATCHES=true -t sogo-unpatched .
```

You can also apply a subset of patches by mounting only the ones you want into `patches/` before building.

### Patch List

The following patches are applied to the SOGo source at build time. They are located in `patches/` and applied in sorted order.

### `01-ldap-multi-password.patch`

**Files:** `SOGoSource.h`, `LDAPSource.m`, `SQLSource.m`, `SOGoUserManager.m`

Enables CalDAV/CardDAV authentication to continue working when the LDAP server is down, and supports accounts with multiple valid LDAP passwords.

- **Always tries LDAP first.** The cache is only used as a fallback when the LDAP server is unreachable — never when it's up.
- **Caches all valid password hashes.** On successful LDAP auth, the SHA1 hash is appended to a `passwords` array in the memcached user attributes. Multiple valid passwords accumulate over time.
- **Offline fallback.** When LDAP is unreachable (detected via connection exceptions), authentication falls back to checking the cached `passwords` array.
- **Invalidates stale passwords.** When LDAP is up and rejects a password that was previously cached, that hash is removed from the cache.
- **Password changes.** When a user explicitly changes their password via SOGo, the old hashes are replaced with the new one.
- Backward-compatible `password` key is maintained alongside the new `passwords` array.

### `02-oidc-refresh-token.patch`

**Files:** `GCSOpenIdFolder.h`, `GCSOpenIdFolder.m`, `SOGoOpenIdSession.m`

Adds OpenID Connect refresh token handling.

### `03-reverse-proxy-cookie.patch`

**Files:** `SOGoSystemDefaults.h`, `SOGoSystemDefaults.m`, `SOGoWebAuthenticator.m`

Adds support for reverse proxy cookie-based authentication.

### `04-oidc-cors-ajax-redirect.patch`

**Files:** `SOGoRootPage.m`, `UIxPageFrame.wox`, `Common.app.js`

Handles CORS and AJAX redirect issues with OIDC authentication flows.

### `05-dav-source-down-503.patch`

**Files:** `SOGoUserManager.h`, `SOGoUserManager.m`, `SOGoDAVAuthenticator.m`

Stops CalDAV/CardDAV from reporting an authentication failure when the cause is an unreachable authentication backend.

- A `401 Unauthorized` tells a DAV client its credentials are wrong. Thunderbird, DAVx5, Evolution and the Apple clients react by re-prompting the user or disabling the account, so an LDAP outage turns into broken clients and support tickets.
- `SOGoUserManager` now records, per request thread, whether the last `-checkLogin:` failed because the source was unreachable rather than because the password was rejected. This builds on the `sourceDown` plumbing added by `01-ldap-multi-password.patch` and only trips after the offline cache fallback has also failed.
- `SOGoDAVAuthenticator` overrides `-preprocessCredentialsInContext:` and `-renderException:inContext:` to turn that case into `503 Service Unavailable` with `Retry-After: 60` and no `WWW-Authenticate` header. DAV clients back off and retry instead of discarding credentials.
- A successful offline cache hit still authenticates normally, and a genuinely wrong password still returns `401`.
- Scoped to DAV and ActiveSync only; the web UI login is untouched.

**Depends on `01-ldap-multi-password.patch`** and must be applied after it.

## Future Plans

Future plans would mainly be having a Helm Chart to deploy this Docker container. Any automation or scripts that will make it easier to build the Helm Chart would be part of the plans. Environment Variables for configuration is also in consideration.

## Contributing

Contributions are welcome! If you have any ideas, feature requests, or bug reports, feel free to open an issue or submit a pull request.
