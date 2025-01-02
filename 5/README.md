# Day 05

<h3 id="ex00">Exercise 00: Toys on a Tree</h3>

After some time, you two put together a structure for a [Binary tree](https://en.wikipedia.org/wiki/Binary_tree) node:

```go
type TreeNode struct {
    HasToy bool
    Left *TreeNode
    Right *TreeNode
}
```

&mdash; Looks like you are supposed to... "hang toys" on trees? - Lily looked a bit confused. - Okay, anyway, let's hope just a boolean value will suffice. But they say it's also wrong to put more toys on one side, should it be uniform?

&mdash; Okay, I get it - you said. - Let's write a function `areToysBalanced` which will receive a pointer to a tree root as an argument. The point is to spit out a true/false boolean value depending on if left subtree has the same amount of toys as the right one. The value on the root itself can be ignored.

So, your function should return `true` for such trees (0/1 represent false/true, equal amount of 1's on both subtrees):

```
    0
   / \
  0   1
 / \
0   1
```

```
    1
   /  \
  1     0
 / \   / \
1   0 1   1
```

and `false` for such trees (non-equal amount of 1's on both subtrees):

```
  1
 / \
1   0
```

```
  0
 / \
1   0
 \   \
  1   1
```

<h3 id="ex01">Exercise 01: Decorating</h3>

&mdash; So, now about this "garland"... It is supposed to be "reeled up" on a tree.

Lily rotated hologram back and forth, trying to think of something. Then suddenly she lights up with enthusiasm.

&mdash; I get it! Let's do it like this... - she draws something that resembles a 3d snake on top of the tree.

So, now you have to write another function called `unrollGarl****and()`, which also receives a pointer to a root node. The idea is to go top down, layer by layer, going right on even horisontal layers and going left on every odd. The returned value of this function should be a slice of bools. So, for this tree:

```
    1
   /  \
  1     0
 / \   / \
1   0 1   1
```

The answer will be [true, true, false, true, true, false, true] (root is true, then on second level we go from left to right, and then on third from right to left, like a zig-zag).

<h3 id="ex02">Exercise 02: Heap of Presents</h3>

&mdash; Perfect! I have no idea what those old dudes meant by "Christmas tree", but I think we've met the general requirements.

&mdash; So, about those "presents"...

&mdash; Presents, right! - Lily raises her elegant finger with a very long purple nail. It was specifically reinforced to fight enemies and (a lot more frequently) to unscrew various devices. - So, let's think of it as a pile. Every such "present" may look like this:

```go
type Present struct {
    Value int
    Size int
}
```

&mdash; Hmm, what's "Value"?

&mdash; Well, some things you tend to value more than the others, right? So they should be comparable.

&mdash; Okay, and "Size" is about how long will it take me to download it, right?

&mdash; Exactly! So, the the coolest things should be on top.

You need to implement a PresentHeap data structure (using built-in library "container/heap" is recommended, but is not strictly required). Presents are compared by Value first (most valuable present goes on top of the heap). *Only* in case two presents have an equal Value, the smaller one is considered to be "cooler" than the other one (wins in comparison).

Apart from the structure itself, you should implement a function `getNCoolestPresents()`, that, given an unsorted slice of Presents and an integer `n`, will return a sorted slice (desc) of the "coolest" ones from the list. It should use the PresentHeap data structure inside and return an error if `n` is larger than the size of the slice or is negative.

So, if we represent each Present by a tuple of two numbers (Value, Size), then for this input:

```
(5, 1)
(4, 5)
(3, 1)
(5, 2)
```

the two "coolest" Presents would be [(5, 1), (5, 2)], because the first one has the smaller size of those two with Value = 5.

<h3 id="ex03">Exercise 03: Knapsack</h3>

&mdash; Wait! - you said. - But how do I know that all these amazing presents won't eat up all the space on my hard drive?

Lily thought for a moment, but then proposed:

&mdash; For this case, let's only download the most valuable presents!

&mdash; But the Heap is using a different ordering and won't help us here...

&mdash; True, true. Anyway, there should be some argument to figure out how to get the most value out of the space you have, right?

...It's been a great winter night in CyberCity. Even though traditions changed a lot in last centuries, you two had a feeling you did everything right. Also, Lily didn't know yet about a cool new portable cyberdeck you've prepared as a gift to her this evening. And you had no idea what's in that small mysterious box on her table.

As a last task, you have to implement a classic dynamic programming algorithm, also known as "Knapsack Problem". Input is almost the same, as in the last task - you have a slice of Presents, each with Value and Size, but this time you also have a hard drive with a limited capacity. So, you have to pick only those presents, that fit into that capacity and maximize the resulting value.

Please write a function `grabPresents()`, that receives a slice of Present instances and a capacity of your hard drive. As an output, this function should give out another slice of Presents, which should have a maximum cumulative Value that you can get with such capacity.
