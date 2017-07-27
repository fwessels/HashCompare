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
|    9216 | 24 | 26 | 28 | 29 |
|   18432 | 25 | 28 | 31 | 28 |
|   36864 | 28 | 33 | 31 | 31 |
|   73728 | 31 | 30 | 33 | 31 |
|  147456 | 36 | 36 | 33 | 34 |
|  294912 | 36 | 36 | 38 | 35 |
|  589824 | 38 | 37 | 40 | 39 |
| 1179648 | 43 | 41 | 40 | 41 |

TODO: add graph

## Conclusion

TODO: add conclusion

## Details

```
h i g h w a y h a s h
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 24 | 11.872592ms
----------18432 |---------- 25 | 18.041524ms
----------36864 |---------- 28 | 47.584637ms
----------73728 |---------- 31 | 108.025128ms
---------147456 |---------- 36 | 218.347766ms
---------294912 |---------- 36 | 479.007373ms
---------589824 |---------- 38 | 1.173577005s
--------1179648 |---------- 43 | 3.139133297s
```

```
b l a k e 2 b
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 26 | 30.5187ms
----------18432 |---------- 28 | 72.318481ms
----------36864 |---------- 33 | 149.28384ms
----------73728 |---------- 30 | 313.881029ms
---------147456 |---------- 36 | 755.909634ms
---------294912 |---------- 36 | 2.040478853s
---------589824 |---------- 37 | 6.130560106s
--------1179648 |---------- 41 | 24.445997385s
```

```
p o l y 1 3 0 5
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 28 | 12.476928ms
----------18432 |---------- 31 | 34.125165ms
----------36864 |---------- 31 | 62.216278ms
----------73728 |---------- 33 | 152.455204ms
---------147456 |---------- 33 | 296.202268ms
---------294912 |---------- 38 | 718.390938ms
---------589824 |---------- 40 | 1.902395985s
--------1179648 |---------- 40 | 5.549007878s
```

```
s i p h a s h
---Permutations |--- Zero bits | Duration
-----------9216 |---------- 29 | 12.950494ms
----------18432 |---------- 28 | 29.749295ms
----------36864 |---------- 31 | 81.086914ms
----------73728 |---------- 31 | 140.770951ms
---------147456 |---------- 34 | 325.780495ms
---------294912 |---------- 35 | 777.761184ms
---------589824 |---------- 39 | 2.568794098s
--------1179648 |---------- 41 | 7.093190263s
```