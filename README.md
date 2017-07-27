# Comparison of Hashing Techniques for Bit Rot detection

## Introduction

At [minio](github.com/minio/minio) we use a hash algorithm in order to detect bit rot on data stored on disk. Thus far we have been using the Blake2B algorithm but recent developments have led us to do new research for a faster algorithm for bit rot protection.

Specifically we have looked into [SipHash](https://131002.net/siphash/) as well as [Poly1305](https://cr.yp.to/mac.html). As a last algorithm we have added [HighwayHash](https://arxiv.org/pdf/1612.06257.pdf) although that is not yet final as of the time of writing.

Obviously Poly1305 is not a "traditional" hash function but since it has nice performance properties and would fit our use case if used in a particular way, we decided to test it out as well.

## Bit rot simulation

Due to the fact that we are interested specifically in bit rot detection, we tried to simulate  bit rot by generating messages of varying lengths and by flipping (XOR-ring) a single bit across the full length of the message and computing the hash for each of these permutations. In addition we also run patterns of 2 up to 8 adjacent bits across each byte of each input message. 

All hashes are sorted and then the smallest difference is searched for between two consecutive hashes. The number of leading zero bits that this smallest difference has then gives us a measure of how close (or how far apart depending on your point of view) we are to a collision with respect to the total number of bits for the respective hash function.

## Results

The following table shows an overview for the number of leading zero bits for the different permutations. 

| Permutations | Highwayhash | Blake2b | Poly1305 | Siphash | 
| -----------:| -------:| -------:| -------:| -------:|
|      9216 | 24 | 26 | 28 | 29 |
|     18432 | 25 | 28 | 31 | 28 |
|     36864 | 28 | 33 | 31 | 31 |
|     73728 | 31 | 30 | 33 | 31 |
|    147456 | 36 | 36 | 33 | 34 |
|    294912 | 36 | 36 | 38 | 35 |
|    589824 | 38 | 37 | 40 | 39 |
|   1179648 | 43 | 41 | 40 | 41 |
|   2359296 | 42 | 44 | 41 | 45 |
|   4718592 | 44 | 48 | 45 | 45 |
|   9437184 | 48 | 46 | 47 | 46 |
|  18874368 | 49 | 46 | 48 | 51 |
|  37748736 | 50 |    |    |    |
|  75497472 | 56 |    |    |    |
| 150994944 | 55 |    |    |    |

Graphically represented this gives the following overview:

![Hash Comparison Overview](https://s3.amazonaws.com/s3git-assets/hash-comparison.png)


## Conclusion

TODO: add conclusion

## Details

```
h i g h w a y h a s h
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 24 | 27.20358ms
----------18432 |---------- 25 | 47.493533ms
----------36864 |---------- 28 | 99.382335ms
----------73728 |---------- 31 | 195.115124ms
---------147456 |---------- 36 | 358.336877ms
---------294912 |---------- 36 | 684.756361ms
---------589824 |---------- 38 | 1.083018009s
--------1179648 |---------- 43 | 2.154209387s
--------2359296 |---------- 42 | 4.001445349s
--------4718592 |---------- 44 | 9.211455489s
--------9437184 |---------- 48 | 23.193931377s
-------18874368 |---------- 49 | 1m7.186158573s
-------37748736 |---------- 50 | 3m43.992417371s
-------75497472 |---------- 56 | 22m14.036457753s
------150994944 |---------- 55 | 2h29m37.881276513s
```

```
b l a k e 2 b
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 26 | 71.118327ms
----------18432 |---------- 28 | 137.560272ms
----------36864 |---------- 33 | 264.832668ms
----------73728 |---------- 30 | 494.631615ms
---------147456 |---------- 36 | 860.939279ms
---------294912 |---------- 36 | 1.62075065s
---------589824 |---------- 37 | 2.866390918s
--------1179648 |---------- 41 | 5.651852469s
--------2359296 |---------- 44 | 12.496053627s
--------4718592 |---------- 48 | 35.131431064s
--------9437184 |---------- 46 | 1m46.990316566s
-------18874368 |---------- 46 | 6m11.394899613s
```

```
p o l y 1 3 0 5
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 28 | 37.327153ms
----------18432 |---------- 31 | 66.769716ms
----------36864 |---------- 31 | 142.288325ms
----------73728 |---------- 33 | 295.353311ms
---------147456 |---------- 33 | 496.196055ms
---------294912 |---------- 38 | 762.110088ms
---------589824 |---------- 40 | 1.475377302s
--------1179648 |---------- 40 | 2.845713401s
--------2359296 |---------- 41 | 5.287880056s
--------4718592 |---------- 45 | 12.390278439s
--------9437184 |---------- 47 | 32.551548229s
-------18874368 |---------- 48 | 1m39.03909055s
```

```
s i p h a s h
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 29 | 32.706182ms
----------18432 |---------- 28 | 58.677561ms
----------36864 |---------- 31 | 114.053811ms
----------73728 |---------- 31 | 232.684029ms
---------147456 |---------- 34 | 463.876011ms
---------294912 |---------- 35 | 829.982109ms
---------589824 |---------- 39 | 1.591968018s
--------1179648 |---------- 41 | 2.685356959s
--------2359296 |---------- 45 | 5.939137835s
--------4718592 |---------- 45 | 14.937767663s
--------9437184 |---------- 46 | 43.330472298s
-------18874368 |---------- 51 | 2m23.595984356s
```
