from datetime import date
from google.appengine.ext import ndb

class Entry(ndb.Model):
	date = ndb.DateProperty(indexed=True)
	count = ndb.IntegerProperty(default=0)

class Counter(ndb.Model):
	name = ndb.StringProperty()
	hash = ndb.StringProperty(indexed=True)
	created_dt = ndb.DateTimeProperty(auto_now_add=True)
	entries = ndb.StructuredProperty(Entry, repeated=True)

	@ndb.transactional
	def increment(self):
	    """Increment the value for a given counter."""
	    # TODO
	    self.count += 1
	    self.put()