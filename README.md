# Simple Page In Go
This project is an attempt to implement a simple page. A page is a fixed-length contiguous block of virtual memory. I hope to be able to use this "page" to implement a very simple database. As you might already know, reading to and writing from disk is much slower than RAM.
Databases read and write data as blocks(and not single row or cell), the page here is going to represent a block in the memory. It will be loaded from disk, modified, and persisted again.

This page support adding, removing, searching and fetching of "cells". Each cell is an array of bytes wrapped by a simple frame.
