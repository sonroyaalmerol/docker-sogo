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
  - `latest` (will always follow the latest container revision of the latest SOGo version)
  - `${SOGo-version}` (e.g. `5.10.0`)
  - `${SOGo-version}-${Container-Revision}` (e.g. `5.10.0-1`)

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

When the container initializes, a script will merge both YAML files to a single YAML file with `yq` and generate the final `/etc/sogo/sogo.conf` file.

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

## Future Plans

Future plans would mainly be having a Helm Chart to deploy this Docker container. Any automation or scripts that will make it easier to build the Helm Chart would be part of the plans. Environment Variables for configuration is also in consideration.

## Contributing

Contributions are welcome! If you have any ideas, feature requests, or bug reports, feel free to open an issue or submit a pull request.
