# Release

## How to Create a Release

1. Create a git tag for the new release

   ```sh
   git tag v<MY-VERSION>
   ```

2. Import our GPG signing private key:
   - retrieve the GPG private key password and store it in an environment
     variable:

       ```sh
        export GPG_PASSWORD="$(vault read -field=master-priv-key-password secret/gpg-key-platform)"
       ```

   - import the signing key:

       ```sh
       vault read -field=subkey-signing-priv-key  secret/gpg-key-platform | \
           gpg --batch --pinentry-mode loopback --passphrase "$GPG_PASSWORD" --import
       ```

3. Set the `GITHUB_TOKEN` environment variable:

   ```sh
   export GITHUB_TOKEN=<MY-TOKEN>
   ```
 
4. Run goreleaser

    ```sh
    goreleaser release
    ```
6. Review the created draft release on
   [GitHub](https://github.com/simplesurance/dependencies-tool/releases) and
   publish it.
   
7. If the release was published, push the local git-tag:

    ```sh
    git push --tags
    ```
  If the release was discarded, delete the local git tag.
