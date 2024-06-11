# GOOGLE <-> CloudFlare group syncer

This app sync particular or regexp google groups to defined or generated CloudFlare Zero Trus Team Lists.

## Build

```bash
make clean && make
```

## SetUp

1) Create Google Cloud Console Service Account. Generage api token in json format. Provide Domain Wide Delegations for that user in admin.google.com

2) Create or use user who has admin(doubtfull, could check) rights in GCP to impersonate.

3) Create CloudFalre Token `https://dash.cloudflare.com/profile/api-tokens`

4) Find the Account ID. Log in to your dash and see the URL `https://dash.cloudflare.com/this>>123132dsdsfd132132sdfsf<<this`

5) Set enviornment variables:

```bash
    export CF_API_KEY=12313212
    export CF_API_EMAIL=account_email@example.com
    export CF_API_ACCOUNTID=123132dsdsfd132132sdfsf
    export GOOGLE_CREDENTIALS= PUT HERE THE CONTENT OF credential.json # IF YOU WANT YOU COULD create google.json in the workdir instead it has less priority
    export GOOGLE_DOMAIN=example.com
```

## RUN

There are 3 modes:

1) Get 1 or more Google Groups and sync them to 1 particular CloudFlare Team List.
In that case target list should be created in advance.

```bash
./dist/google-cloudflare-sync -google_groups user1@example.com,user2@example.com -google_impersonate admin_user@example.com -cf_list_name target_list_in_cf
```

2) Set the mask for searching groups in Google and app will find them and create or update existsing ones in CF.

```bash
./dist/google-cloudflare-sync -google_groups_regex 'email:cflist*' -google_impersonate admin_user@example.com
```

3) Delete stale users. Set google groups where active users are defined. App will bisect this groups with active CF users and deactivate CF users not listed in defined groups. Complete deletion is not possible due to API restrictions.
`There is currently no way to delete or archive a user record. You can remove a user from a seat, but their user record will remain in Zero Trust.`

```bash
./dist/google-cloudflare-sync -delete_stale -google_groups "internal-users@example.com,external-users@example.com" -google_impersonate admin_user@example.com
```

## DEBUG

Use `-debug` key
