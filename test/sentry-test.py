import sentry_sdk
import random
from sentry_sdk import capture_exception
from sentry_sdk import capture_message
from sentry_sdk import add_breadcrumb


n = random.choice([1, 1, 1, 1, 1, 2, 2, 3])

sentry_sdk.init(
    dsn="http://gtn-py@localhost:8000/1",
    # Enable performance monitoring
    enable_tracing=True,
    send_default_pii=True,
    environment=random.choice(["development", "staging", "production"]),
    release=f"myapp@0.0.{n}"
)

add_breadcrumb(
    category='auth',
    message='Authenticated user %s' % 'hexylena',
    level='info',
)

sentry_sdk.set_context("character", {
    "name": "Mighty Fighter",
    "age": 19,
    "attack_type": "melee"
})


def b():
    a_potentially_failing_function()

def a():
    b()

try:
    a()
except Exception as e:
    # Alternatively the argument can be omitted
    capture_exception(e)


# capture_message('Something went wrong')
