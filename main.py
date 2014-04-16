from flask import Flask, render_template, jsonify, request, abort
from slugify import slugify
from hashlib import sha256
from models import *
import os

app = Flask(__name__)

@app.route('/')
def index():
    """Return a friendly HTTP greeting."""
    hash = sha256(os.urandom(1024)).hexdigest()
    return render_template('index.html', hash=hash)

@app.route('/<hash>')
def account(hash):
    """View a list of counters"""
    return render_template('account.html')

@app.route('/<hash>/counters', methods=['GET', 'POST'])
def counter_list(hash):
	"""List counters"""
	counters = []
	results = Counter.query()

	# Create new counter
	if request.method == 'POST':
		counter = Counter(name=request.form.get('name'), id=slugify(request.form.get('name')))
		if counter.name is None:
			abort('name is required')
		counter.put()
		counters.append({
			"id": counter.key.id(),
			"name": counter.name,
			"created_dt": counter.created_dt.strftime('%Y-%m-%d %H:%M:%S'),
			"entries": []
		})
	
	# Return list of counters
	for counter in results:
		output = {
			"id": counter.key.id(),
			"name": counter.name,
			"created_dt": counter.created_dt.strftime('%Y-%m-%d %H:%M:%S'),
			"entries": []
		}
		for entry in counter.entries:
			output['entries'].append({
				"date": entry.date,
				"count": entry.count
			})
		counters.append(output)
	return jsonify(counters=counters)

@app.route('/<hash>/counters/<counter_id>', methods=['POST'])
def counter(hash, counter_id):
	counter = ndb.Key(Counter, counter_id).get()
	if counter is None:
		abort('Could not find entity')
	return jsonify(counter)

@app.errorhandler(404)
def page_not_found(e):
    """Return a custom 404 error."""
    return '404 Not Found', 404

@app.errorhandler(500)
def page_not_found(e):
    """Return a custom 500 error."""
    return 'Sorry, unexpected error: {}'.format(e), 500
