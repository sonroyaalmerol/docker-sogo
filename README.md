# SOGo Docker Container
[![Release](https://img.shields.io/github/v/release/sonroyaalmerol/docker-sogo)](https://github.com/sonroyaalmerol/docker-sogo/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/sonroyaalmerol/docker-sogo)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo)
[![Docker Stars](https://img.shields.io/docker/stars/sonroyaalmerol/docker-sogo)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo)
[![Image Size](https://img.shields.io/docker/image-size/sonroyaalmerol/docker-sogo/latest)](https://hub.docker.com/r/sonroyaalmerol/docker-sogo/tags)
[![Release](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml/badge.svg)](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml)
[![Release](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml/badge.svg)](https://github.com/sonroyaalmerol/docker-sogo/actions/workflows/release.yml)

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

- **Automated Builds**: SOGo releases are checked every day for updates and will trigger the Docker image build automatically.
- **Follows the official SOGo version tagging**: Docker image tags follows the version tagging of SOGo releases.
- **YAML Configuration Support**: Users can define SOGo configurations using YAML files, allowing for easier management and version control. The included script automatically merges and converts YAML configurations into the required OpenStep plist format, simplifying the setup process.

## Usage

### Running the Docker Container

Assuming you have your YAML configuration files ready, you can run the Docker container using the following command:

```bash
docker run -v /path/to/sogo/configs:/etc/sogo/sogo.conf.d/ sonroyaalmerol/docker-sogo
```

Replace `/path/to/sogo/configs` with the directory containing your YAML configuration files.

The NGINX process inside the container will host the web server on port 80 by default. See the [default NGINX config](https://github.com/sonroyaalmerol/docker-sogo/blob/main/default-configs/nginx.conf) for more info.

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