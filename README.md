BitX
========
Status: Early development and design.

Next Goal: A simple demo of a usable applications, probably file backup and sharing.

Needs:

* A tree hash format, will be expansive to change once deployed. Partially Coded, using sha256.

* Networking:
   * Data serialization, using Protocal Buffers.
   * Transport layer, UDP, messages are designed to be small and independent, Server is reactive to requests. Connections, user sessions, and transport layer securities will be considered later.

* Collections:
   * A common formate for a collection of links.
   * Can be used to implement folders.
   * Can be understood by servers help to implement many features. Such as *Retain*.

* Delta Update:
   * An efficiency improvement for updates of large files.
   * Unlike tree hash, multiple algrithims can co-exist with low over-head. The client simply choose the best of provided obtions by algrithm and delta size.
   * After application of delta, the content is checked against the orignal hash. A malicious delta suggestion will delay downloading, but can not modify contents.

* Public-key Infrastructure
   * For the first demo, a pre-configered public key to validate updates.


Idea
-----

To build a distributed data cache.

* Static data is referenced by hash+length. Large data blocks uses tree hash to aid in multi-source download and error recovery. The tree structure is unique per length.
* Dynamic data is referenced by public key, data with a higher version number replaces older data. This update channel is protected by securing the private key.
* Such that any data can be verified using the hashes and keys, allowing anyone to host.

Then different applications can be build using those services.


Basic use cases should be easy
----------

####To upload data:

1. Client makes data available for download, for a set of servers.
2. Client sends *Retain* requests to servers, with the link to data.
3. Servers download data from client, reusing connection established by *Retain*. Servers can also download from each other.
4. Client can measure redundancy by number of copies on servers.

If the set of servers is inclusive enough, then only #1 is needed. *Retain* only improves redundancy.

Servers have a *Retain* set. Servers should provide data in this set without external dependence. Servers may also provide other contents by cache or proxy without the same guarantees, allow freeing of non-retained content.

Clients should control access by encrypting content before upload. Servers can also limit access, mostly to control resource usage and costs.


####To download data:

1. Client request data by link from any server, magic.
2. Client can prefetch/offline data by *Retain* in a local server.

To free application development from complexities, most clients should use a daemon or library, which implements features of servers.

Servers in consumer devices on a wireless network and server in data centers does not need the same behaviour. They would take different steps when relaying a download request. The consumer device can simply relay all requests to another server in a data center, geographically close, to do all the hard work. A list of fail over servers will provide redundancy. For hard work, see [Link Magic](#linkMagic).


####Real time updates of dynamic data:

Server push, nothing new, right?


Link Magic<a id="linkMagic"></a>
-------
A link must uniquely identify content by hash+length or public key.
But looking for global resources this way is expansive, short-cuts to locate data should be taken whenever possible.

* When data is first uploaded, the client can generate links with location information, which identifies the client and the servers that *Retain* the data. This gives fast data retrieval in most use cases.
* Available locations can change overtime, this necessities indexing services. There are many existing methods that provide this function, a choice is to be determined.


Complementary Ideas
-----------

#### Reverse link lookups
Allows searching for things that link to the current content. With tags to identify the type of relationship, such as: reference, revision, or commentary. This enables new metadata to be added to existing static content. Some sort of reputation rating is needed to combat spam.

#### Public-key Infrastructure
Just as the trust of static hash+length is derived from the trust of the source of the reference link. The trust of a public key to maintain an update channel is assured by its source.

Well a public key only needs to be protected against modification, a private key attemps to hide from attackers and be easily accessible to publish new content. We can not make the assumption that a private key will never be lost or stolen. The system needs to delegates authorization to more limited keys, conforms revocation statues in a timely manner, and allows the requirement of multi-party sign-offs for critical functions.

#### Resource Allocation
A service provider decide who is allowed to use its resources and how much, based on commercial or community agreements. This means clients also needs to manage acounts with *Retains* and download capacity. The same relationship applies between providers to exchanging data.

Manual management of individual acounts decrease the usablity of a large distributed system. A system that automatical choose services depending on advertised price, stability, and responsiveness; and ratings to attest to their truth, will be helpful.

Application developers have an option to choose how acounts are founded by ether requesting the user to provide access or using the developer's. Allowing ether party to pay for infostructure costs.



License
=====
2-Clause BSD, see LICENSE file.