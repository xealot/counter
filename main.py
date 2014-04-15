from flask import Flask
from hashlib import sha256
import os

app = Flask(__name__)

@app.route('/')
def index():
    """Return a friendly HTTP greeting."""
    hash = sha256(os.urandom(1024)).hexdigest()
    return '<a href="/{hash}">New account</a>'.format(hash=hash)

@app.route('/<hash>')
def account(hash):
    """View a list of counters"""
    return 'Save this page. This is your new account, accessible to anyone with the url.'


@app.errorhandler(404)
def page_not_found(e):
    """Return a custom 404 error."""
    return '404 Not Found', 404


@app.errorhandler(500)
def page_not_found(e):
    """Return a custom 500 error."""
    return 'Sorry, unexpected error: {}'.format(e), 500
