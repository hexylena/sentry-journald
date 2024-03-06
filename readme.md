# Sentry-Journald

Do you like sentry's data collection, but don't want to run a complicated sentry server? Do you still want to see those cute lil error messages?
Boy howdy do I have the 200 lines of rubbish Golang code for you!

```
$ journalctl -f # optionally -t sentry
Mar 06 13:13:22 w-galaxy sentry[467887]: [event] (proj=web-demo env=production) [http://localhost:4001/test.html:38:5] [http://localhost:4001/test.html:1:1] ReferenceError: someOtherFunction is not defined
Mar 06 13:13:51 w-galaxy sentry[467887]: [event] (proj=py-demo env=development) [sentry-test.py:29:0] NameError: name 'a_potentially_failing_function' is not defined
Mar 06 13:13:51 w-galaxy sentry[467887]: [event] (proj=py-demo env=development) Something went wrong
```

If you want the full data it's stuffed into additional fields in the `journald` json output.

```json
{
  "MESSAGE": "[event] (proj=web-demo env=production) [http://localhost:4001/test.html:38:5] [http://localhost:4001/test.html:1:1] ReferenceError: someOtherFunction is not defined",
  "MESSAGE_ID": "2b5238a100474170b7bb6bd78fc8842f",
  "PRIORITY": "3",
  "PROJECT_ID": "1",
  "REMOTE_ADDR": "127.0.0.1:57516",
  "REQUEST_HEADERS": "{\"User-Agent\":\"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0\"}",
  "REQUEST_METHOD": "POST",
  "REQUEST_REMOTE_ADDR": "127.0.0.1:57516",
  "REQUEST_URL": "http://localhost:4001/test.html",
  "SENTRY_CLIENT": "sentry.javascript.browser/7.105.0",
  "SENTRY_CONTEXTS": "{\"trace\":{\"span_id\":\"a091418210b527fc\",\"trace_id\":\"64c9cb5312174c04be784ebdd66d094e\"}}",
  "SENTRY_DIST": "my-project-name@2.3.12",
  "SENTRY_ENVIRONMENT": "production",
  "SENTRY_KEY": "password",
  "SENTRY_PLATFORM": "javascript",
  "SENTRY_RELEASE": "my-project-name@2.3.12",                                                                                                                                                                   "SENTRY_SERVER_NAME": "",
  "SENTRY_TIMESTAMP": "1.709726111276e+09",
  "SENTRY_VERSION": "7",
  "SYSLOG_IDENTIFIER": "sentry",
},
```

## Configuration

There is a `--port` flag (also available as `PORT`) to configure a port from the environment.

Just like real sentry there is no actual authentication of the user, so, it is up to you to add any filtering and anti-spam/abuse you want.

## systemd

A unit file is included for your convenience.

## License

EUPL-1.2 (it's like agpl! but european flavour.)
