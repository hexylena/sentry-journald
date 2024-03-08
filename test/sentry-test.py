import sentry_sdk
import random
from sentry_sdk import capture_exception
from sentry_sdk import capture_message
from sentry_sdk import add_breadcrumb
from sentry_sdk import set_user


n = random.choice([1, 1, 1, 1, 1, 2, 2, 3])

sentry_sdk.init(
    dsn="http://my-python-project@localhost:8000/1",
    # Enable performance monitoring
    enable_tracing=True,
    send_default_pii=True,
    environment=random.choice(["development", "staging", "production"]),
    release=f"myapp@0.0.{n}"
)

# These are not currently supported (PRs welcome!)
add_breadcrumb(
    category='auth',
    message='Authenticated user %s' % 'hexylena',
    level='info',
)
# Sent as:
#  "breadcrumbs": {
#    "values": [
#      {
#        "category": "auth",
#        "level": "info",
#        "message": "Authenticated user hexylena",
#        "timestamp": "2024-03-08T10:35:49.573714Z",
#        "type": "default"
#      }
#    ]
#  },


# This is though!
sentry_sdk.set_context("user", "hexylena")
sentry_sdk.set_context("level", "info")
# Sent as:
#  "contexts": {
#    "level": "info",
#    "runtime": {
#      "build": "3.11.7 (main, Dec 18 2023, 00:00:00) [GCC 13.2.1 20231011 (Red Hat 13.2.1-4)]",
#      "name": "CPython",
#      "version": "3.11.7"
#    },
#    "trace": {
#      "parent_span_id": null,
#      "span_id": "9f1128787d8b8323",
#      "trace_id": "b6f4d1a8eff0439b800644e067013f19"
#    },
#    "user": "hexylena"
#  },


# Not this (yet)
# set_user({"username": "hexylena", "email": "jane.doe@example.com"})
# Sent as:
#  "user": {
#    "email": "jane.doe@example.com",
#    "username": "hexylena"
#  }



def b():
    a_certainly_failing_function()

def a():
    b()

try:
    a()
except Exception as e:
    # Alternatively the argument can be omitted
    capture_exception(e)


# capture_message('Something went wrong')
