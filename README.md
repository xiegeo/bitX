BitX
========
Status: Early development and design.

Next Goal: A simple demo of a usable applications, probably file backup and sharing.

Needs:

* A tree hash format, will be expensive to change once deployed. Partially Coded, using sha256.

* Networking:
	* Data serialization, using Protocal Buffers. See [network.proto](src\network\network.proto)
	* Transport layer, UDP, messages are designed to be small and independent, Server is reactive to requests. Connections, user sessions, and transport layer securities will be considered later.

* Collections:
	* A common format for a collection of links.
	* Can be used to implement folders.
	* Can be understood by servers to help implement features based on lists, such as *Retain*.

* Delta Update:
	* An efficiency improvement for updates of large files.
	* Unlike tree hash, multiple algorithms can co-exist with low overhead. The client simply choose the best of provided options by algorithm and delta size.
	* After application of delta, the content is checked against the original hash. A malicious delta suggestion will delay downloading, but can not modify contents.

* Public-key Infrastructure
	* For the first demo, a pre-configured public key to validate updates.


Idea
-----

To build a distributed data cache.

* Static data is referenced by hash+length. Large data blocks uses tree hash to aid in multi-source download and error recovery. The tree structure determined by length.
* Dynamic data is referenced by public key, data with a higher version number replaces older data. This update channel is protected by securing the private key.
* There for, any data can be verified using the hashes and keys, allowing anyone to distribute data.

Then different applications can be build using those services.


Basic use cases should be easy
----------

####To upload data:

1. Client, acking like a server, makes data available for download.
2. Client sends *Retain* requests to servers, with the link to data, including hash+length and addresses of client and other retaining servers.
3. Servers download data from client and each other, reuse use of connection established by client *Retain* request to work around NAT and firewall issues.
4. Client can control redundancy by changing the distribution of retaining servers.

The client is ready to publish a link for public consumbsion of data at step 1. Everything else is for redundancy.

Servers have a *Retain* set. Servers should provide data in this set without external dependence. Servers may also provide other contents by cache or proxy without the same guarantees, allow freeing of non-retained content.

Clients should control access by encrypting content before upload. Servers can also limit access, mostly to control resource usage and costs.


####To download data:

1. Client can request data by link from any server.
2. Client can prefetch/offline data by *Retain* in a local server.

To free application development from complexities, most clients should use a daemon or library, which implements features of servers.

Servers in consumer devices on a wireless network and server in data centers does not need the same behaviour. They would take different steps when relaying a download request. The consumer device can simply relay all requests to another server in a data center, geographically close, to do all the hard work. A list of failover servers will provide redundancy. For hard work, see [Link Magic](#link-magic).


####Real time updates of dynamic data:

Server push, nothing new, right?


Link Magic<a name="link-magic"></a>
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

Well a public key only needs to be protected against modification, a private key attempts to hide from attackers and be easily accessible to publish new content. We can not make the assumption that a private key will never be lost or stolen. The system needs to delegates authorization to more limited keys, conforms revocation status in a timely manner, and allows the requirement of multi-party sign-offs for critical functions.

#### Resource Allocation
A service provider decide who is allowed to use its resources and how much, based on commercial or community agreements. This means clients also needs to manage accounts with *Retains* and download capacity. The same relationship applies between providers to exchanging data.

Manual management of individual accounts decrease the usability of a large distributed system. A system that automatically choose services depending on advertised price, stability, and responsiveness; and ratings to attest to their truth, will be helpful.

Application developers have an option to choose how accounts are funded by either requesting the user to provide access or using the developer's. Allowing either party to pay for infrastructure costs.



License
=====
2-Clause BSD, see [LICENSE](LICENSE) file.