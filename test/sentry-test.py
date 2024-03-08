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

# These are not currently supported
add_breadcrumb(
    category='auth',
    message='Authenticated user %s' % 'hexylena',
    level='info',
)

# This is though!
sentry_sdk.set_context("user", "hexylena")
sentry_sdk.set_context("level", "info")

# Also
set_user({"username": "hexylena", "email": "jane.doe@example.com"})


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
