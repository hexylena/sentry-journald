# Sentry-Journald

Do you like sentry's data collection, but don't want to run a complicated sentry server? Do you still want to see those cute lil error messages?
Boy howdy do I have the 200 lines of rubbish Golang code for you!

```
$ journalctl -f # optionally -t sentry
Mar 06 13:13:22 w-galaxy sentry[467887]: [event] (proj=nil env=production) [http://localhost:4001/test.html:38:5] [http://localhost:4001/test.html:1:1] ReferenceError: someOtherFunction is not defined
Mar 06 13:13:51 w-galaxy sentry[467887]: [event] (proj=nil env=development) [sentry-test.py:29:0] NameError: name 'a_potentially_failing_function' is not defined
Mar 06 13:13:51 w-galaxy sentry[467887]: [event] (proj=nil env=development) Something went wrong
```

If you want the full data it's stuffed into additional fields in the `journald` json output.

```json
{
  "MESSAGE": "[event] (proj=nil env=production) [http://localhost:4001/test.html:38:5] [http://localhost:4001/test.html:1:1] ReferenceError: someOtherFunction is not defined",
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
  "SENTRY_RELEASE": "my-project-name@2.3.12",                                                                                                                                                                       "SENTRY_SERVER_NAME": "",
  "SENTRY_TIMESTAMP": "1.709726111276e+09",
  "SENTRY_VERSION": "7",
  "SYSLOG_IDENTIFIER": "sentry",
},
{
  "MESSAGE": "[event] (proj=nil env=development) [sentry-test.py:29:0] NameError: name 'a_potentially_failing_function' is not defined",
  "MESSAGE_ID": "dbe9112bec734cc7b21b1f22c979f747",
  "PRIORITY": "3",
  "PROJECT_ID": "1",
  "REMOTE_ADDR": "[::1]:37282",
  "REQUEST_HEADERS": "null",
  "REQUEST_METHOD": "POST",
  "REQUEST_REMOTE_ADDR": "[::1]:37282",
  "REQUEST_URL": "",
  "SENTRY_CLIENT": "",
  "SENTRY_CONTEXTS": "{\"character\":{\"age\":19,\"attack_type\":\"melee\",\"name\":\"Mighty Fighter\"},\"runtime\":{\"build\":\"3.11.7 (main, Dec 18 2023, 00:00:00) [GCC 13.2.1 20231011 (Red Hat 13.2.1-4)]\",\
"name\":\"CPython\",\"version\":\"3.11.7\"},\"trace\":{\"parent_span_id\":null,\"span_id\":\"90416ef04a2ee00d\",\"trace_id\":\"6f3d1f39dad9475e80b34580e3496611\"}}",
  "SENTRY_DIST": "myapp@0.0.1",
  "SENTRY_ENVIRONMENT": "development",
  "SENTRY_KEY": "",
  "SENTRY_PLATFORM": "python",
  "SENTRY_RELEASE": "myapp@0.0.1",
...}
```

## License

EUPL-1.2 (it's like agpl! but european flavour.)
