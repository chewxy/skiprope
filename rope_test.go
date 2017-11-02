package skiprope

import (
	"bytes"
	"fmt"
	"testing"
	"testing/quick"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func validRope(t *testing.T, r *Rope) {
	assert := assert.New(t)
	assert.Condition(func() bool { return r.Head.height >= 1 }, "Height has to be greater than 1")

	last := r.Head.nexts[r.Head.height-1]
	assert.Equal(r.size, last.skipped, "Expect to have skipped %d for the last element on the skiplist", r.size)
	assert.Nil(last.knot, "Last Knot not nil")

	var runeCount int
	s := skiplist{r: r}
	for i := 0; i < r.Head.height; i++ {
		s.s[i].knot = &r.Head
	}

	for n := &r.Head; n != nil; n = n.nexts[0].knot {
		assert.Condition(func() bool { return n.used > 0 }, "Expected a used count of greater than 0")
		assert.Condition(func() bool { return n.height <= MaxHeight }, "node cannot be greater than MaxHeight - %d", n.height)

		for i := 0; i < n.height; i++ {
			assert.Equal(s.s[i].knot, n, "search[%d] should be %p", i, n)
			assert.Equal(s.s[i].skipped, runeCount, "RuneCount should be the same as skipped.")

			s.s[i].knot = n.nexts[i].knot
			s.s[i].skipped += n.nexts[i].skipped
		}
		runeCount += n.nexts[0].skipped
	}

	for i := 0; i < r.Head.height; i++ {
		assert.Nil(s.s[i].knot)
		assert.Equal(runeCount, s.s[i].skipped)
	}

	assert.Equal(runeCount, r.size)
}

func TestEmptyRope(t *testing.T) {
	r := New()
	assert.Equal(t, "", r.String())
	assert.Equal(t, 0, r.Size())
}

func TestPlay(t *testing.T) {
	// t.Skip()
	r := New()
	if err := r.Insert(0, "0123456789 hello world ab2cdefghi fakk1 eir3d"); err != nil {
		t.Fatal(err)
	}
	rs := r.SubstrRunes(5, 18)
	if string(rs) != "56789 hello w" {
		t.Errorf("Expected %q. Got %q instead", "56789 hello w", string(rs))
	}

	r.InsertRunes(10, []rune("ADDED"))
	rs = r.SubstrRunes(0, r.size)
	if string(rs) != "0123456789ADDED hello world ab2cdefghi fakk1 eir3d" {
		t.Error("Add failed. Got %q instead", string(rs))
	}

	// erase "ADDED"
	if err := r.EraseAt(10, 5); err != nil {
		t.Error(err)
	}
	rs = r.SubstrRunes(0, r.size)
	if string(rs) != "0123456789 hello world ab2cdefghi fakk1 eir3d" {
		t.Errorf("Erase failed. Got %q instead", string(rs))
	}
}

func TestPlay2(t *testing.T) {
	// t.Skip()
	r := New()
	if err := r.Insert(0, "short"); err != nil {
		t.Fatal(err)
	}
	if err := r.Insert(5, "short"); err != nil {
		t.Fatal(err)
	}
	if r.Size() != 10 {
		t.Errorf("Expected size of rope to be 10. Got %d instead", r.Size())
	}
	if r.Runes() != 10 {
		t.Errorf("Expected size of rope to be 10. Got %d instead", r.Runes())
	}
	if r.String() != "shortshort" {
		t.Errorf("Expected rope value to be \"shortshort\". Got %q instead", r.String())
	}

	sss := "this is a super long string of characters inserted into the middle "
	if err := r.Insert(5, sss); err != nil {
		t.Fatal(err)
	}
	if r.String() != "short"+sss+"short" {
		t.Errorf("Insertion failed. Got %q instead", r.String())
	}
	assert.Equal(t, "short"+sss+"short", r.String())
}

func TestQC(t *testing.T) {
	r := New()
	f := func(a string) (ok bool) {
		// insert at 1, because 1 is greater than 0. This tests the resilience of the rope
		if err := r.Insert(1, a); err != nil {
			t.Errorf("Failed to insert during QC: %v", err)
			return false
		}

		// check
		b := r.String()
		if b != a {
			t.Errorf("Expected %q. Got %q instead. Len(a): %d | %d", a, b, len(a), r.size)
			return false
		}

		// del at len(a)+1,  This tests the resilience of the rope
		if err := r.EraseAt(0, len(a)); err != nil {
			t.Errorf("Failed to erase during QC: %v", err)
			return false
		}

		// check
		c := r.String()
		if c != "" {
			t.Errorf("Expected %q. Got %q instead", "", c)
			return false
		}

		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 2000}); err != nil {
		t.Fatal(err)
	}
}

func TestBasic(t *testing.T) {
	r := New()
	if err := r.Insert(0, a); err != nil {
		t.Fatal(err)
	}
	if r.Size() != len(a) {
		t.Errorf("Expected same size of %d. Got %d instead", len(a), r.size)
	}
	if r.runes != len([]rune(a)) {
		t.Errorf("Expected same number of runes %d. Got %d instead", len([]rune(a)), r.runes)
	}

	// Test Indexing
	if roon := r.Index(4); roon != 'e' {
		t.Errorf("Expected rune at 4 to be 'e'. Got %q instead", roon)
	}

	if roon := r.Index(BucketSize); roon != 'n' {
		t.Errorf("Expected rune at %d to be 'n'. Got %q instead", BucketSize, roon)
	}

	if roon := r.Index(BucketSize + 1); roon != ' ' {
		t.Errorf("Expected rune at %d to be 'n'. Got %q instead", BucketSize, roon)
	}

	if roon := r.Index(len([]rune(a))); roon != 0 {
		t.Errorf("Expected rune at %d to be 'n'. Got %q instead", len([]rune(a)), roon)
	}

	if roon := r.Index(2 * len([]rune(a))); roon != -1 {
		t.Errorf("Expected rune at 2*%d to be 'n'. Got %q instead", len([]rune(a)), roon)
	}
}

func TestBefore(t *testing.T) {
	// short
	r := New()
	if err := r.Insert(0, "Hello World"); err != nil {
		t.Fatal(err)
	}

	before, char, err := r.Before(9, unicode.IsSpace)
	if err != nil {
		t.Error(err)
	}

	if before != 6 {
		t.Errorf("Expected before to be 6. Got %d instead", before)
	}
	if char != ' ' {
		t.Errorf("Expected char to be ' '. Got %q instead", char)
	}
}

func TestByteOffset(t *testing.T) {
	a := []byte("hello world")
	offset := byteOffset(a, 5)
	assert.Equal(t, 5, offset)
	assert.Equal(t, ' ', rune(a[offset]), "Expected 'o'. Got %q instead", a[offset])

	// combining diacritics are one separate character
	a = []byte("Sitting in a café, writing this test")
	offset = byteOffset(a, 18)  // 19th character is ',',  and starts at byte offset 20
	assert.Equal(t, 19, offset) // the combining diacritic is 2 characters wide
	assert.Equal(t, ',', rune(a[offset]), "Expected ','. Got %q instead", string(a[offset:]))
}

func TestRope_ByteOffset(t *testing.T) {
	r := New()
	nonUTF8 := "hello world"
	if err := r.Insert(0, nonUTF8); err != nil {
		t.Fatal(err)
	}
	for i := range nonUTF8 {
		b := r.ByteOffset(i)
		assert.Equal(t, i, b)
	}
}

func TestRope_Write(t *testing.T) {
	r := New()
	expected := "Standing at home, writing this test"
	written, err := r.Write([]byte(expected))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, r.String())
	assert.Equal(t, len(expected), written)
}

func TestRope_WriteAppend(t *testing.T) {
	r := New()
	expected := "Sitting at home, writing this test"
	written1, err := r.Write([]byte(expected[:5]))
	if err != nil {
		t.Fatal(err)
	}
	written2, err := r.Write([]byte(expected[5:]))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, r.String())
	assert.Equal(t, len(expected), written1+written2)
}

func ExampleBasic() {
	r := New()
	_ = r.Insert(0, "Hello World. This is a long sentence. The purpose of this long sentence is to make sure there is more than BucketSize worth of runes")

	char := r.Index(70)
	fmt.Printf("Char at 70: %q\n", char)

	l := r.Size()
	fmt.Printf("Length: %d\n", l)

	// Output:
	// Char at 70: 'e'
	// Length: 132
	//
}

func ExampleRope_SubstrBytes_Multibyte() {
	r := New()
	_ = r.Insert(0, "你好world")
	// The string is a 11 byte string. Both chinese characters are 3 bytes each.

	// The SubstrBytes method indexes by runes.
	// if we want only "好", we'd have to call SubstrBytes(1, 2),
	// not SubstrBytes(3,6), which is what you'd do if it indexes by bytes.
	bytes := r.SubstrBytes(1, 2)
	fmt.Println(string(bytes))
	// Output:
	// 好
}

func ExampleRope_Before() {
	r := New()
	_ = r.Insert(0, "Hello World. This is a long sentence. The purpose of this long sentence is to make sure there is more than BucketSize worth of runes")
	before, char, err := r.Before(70, unicode.IsSpace)
	if err != nil {
		fmt.Printf("Error: cannot get before 70: %v", err)
	}

	fmt.Printf("First whitespace before position 70: %d - %q\n", before, char)

	// Output:
	// First whitespace before position 70: 63 - ' '
}

func ExampleRope_ByteOffset() {
	r := New()
	_ = r.Insert(0, "你好world")
	b0 := r.ByteOffset(1)     // 1st rune is '好' - 3
	b1 := r.ByteOffset(2)     // 2nd rune is 'w' - 6
	bErr := r.ByteOffset(200) // impossible

	fmt.Printf("b0: %d\nb1: %d\nbErr: %d", b0, b1, bErr)

	// Output:
	// b0: 3
	// b1: 6
	// bErr: -1
}

// func TestLines(t *testing.T) {
// 	r := New()
// 	s := `1 Hello World. The first line is intentionally super long as to make sure there are several blocks
// 	2
// 	3 你好世界
// 	4 Last line!
// 	`
// 	if err := r.Insert(0, s); err != nil {
// 		t.Fatal(err)
// 	}
// 	row, col := r.RowCol(100) // 2
// 	if row != 2 {
// 		t.Errorf("Expected row to be 2. Got %d instead", row)
// 	}
// 	if col != 1 {
// 		t.Errorf("Expected col to be 1. Got %d instead", col)
// 	}

// 	row, col = r.RowCol(103) // 3
// 	if row != 3 {
// 		t.Errorf("Expected row to be 3. Got %d instead", row)
// 	}
// 	if col != 1 {
// 		t.Errorf("Expected col to be 1. Got %d instead", col)
// 	}

// 	s = `1 Hello World. The first line is intentionally super long as to make sure there are several blocks
// 	2
// 	3 你好世界. Here is another long வாக்கியம் to make sure that the bucket is exceeded
// 	4 Last line!
// 	`
// 	r = New()
// 	if err := r.Insert(0, s); err != nil {
// 		t.Fatal(err)
// 	}
// 	row, col = r.RowCol(185) // 4
// 	if row != 4 {
// 		t.Errorf("Expected row to be 4. Got %d instead", row)
// 	}
// 	if col != 1 {
// 		t.Errorf("Expected col to be 1. Got %d instead", col)
// 	}
// }

// func TestSkiplist_find(t *testing.T) {
// 	r := New()
// 	s := skiplist{r: r}
// 	k, offset, err := s.find(0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if k != &r.Head {
// 		t.Error("Expected knot at 0 to be head")
// 	}

// 	r.Insert(0, "0123456789 hello world ab2cdefghi fakk1 eir3d")
// 	if k, offset, err = s.find(18); err != nil {
// 		t.Error(err)
// 	}
// 	if offset != 2 {
// 		t.Error("Expected offset to be 2")
// 	}
// }

const a = `Alice was beginning to get very tired of sitting by her sister on the bank, and of having nothing to do: once or twice she had peeped into the book her sister was reading, but it had no pictures or conversations in it, ‘and what is the use of a book,’ thought Alice ‘without pictures or conversations?’

So she was considering in her own mind (as well as she could, for the hot day made her feel very sleepy and stupid), whether the pleasure of making a daisy-chain would be worth the trouble of getting up and picking the daisies, when suddenly a White Rabbit with pink eyes ran close by her.

There was nothing so very remarkable in that; nor did Alice think it so very much out of the way to hear the Rabbit say to itself, ‘Oh dear! Oh dear! I shall be late!’ (when she thought it over afterwards, it occurred to her that she ought to have wondered at this, but at the time it all seemed quite natural); but when the Rabbit actually took a watch out of its waistcoat-pocket, and looked at it, and then hurried on, Alice started to her feet, for it flashed across her mind that she had never before seen a rabbit with either a waistcoat-pocket, or a watch to take out of it, and burning with curiosity, she ran across the field after it, and fortunately was just in time to see it pop down a large rabbit-hole under the hedge.

In another moment down went Alice after it, never once considering how in the world she was to get out again.

The rabbit-hole went straight on like a tunnel for some way, and then dipped suddenly down, so suddenly that Alice had not a moment to think about stopping herself before she found herself falling down a very deep well.

Either the well was very deep, or she fell very slowly, for she had plenty of time as she went down to look about her and to wonder what was going to happen next. First, she tried to look down and make out what she was coming to, but it was too dark to see anything; then she looked at the sides of the well, and noticed that they were filled with cupboards and book-shelves; here and there she saw maps and pictures hung upon pegs. She took down a jar from one of the shelves as she passed; it was labelled ‘ORANGE MARMALADE’, but to her great disappointment it was empty: she did not like to drop the jar for fear of killing somebody, so managed to put it into one of the cupboards as she fell past it.

‘Well!’ thought Alice to herself, ‘after such a fall as this, I shall think nothing of tumbling down stairs! How brave they’ll all think me at home! Why, I wouldn’t say anything about it, even if I fell off the top of the house!’ (Which was very likely true.)

Down, down, down. Would the fall never come to an end! ‘I wonder how many miles I’ve fallen by this time?’ she said aloud. ‘I must be getting somewhere near the centre of the earth. Let me see: that would be four thousand miles down, I think—’ (for, you see, Alice had learnt several things of this sort in her lessons in the schoolroom, and though this was not a very good opportunity for showing off her knowledge, as there was no one to listen to her, still it was good practice to say it over) ‘—yes, that’s about the right distance—but then I wonder what Latitude or Longitude I’ve got to?’ (Alice had no idea what Latitude was, or Longitude either, but thought they were nice grand words to say.)

Presently she began again. ‘I wonder if I shall fall right through the earth! How funny it’ll seem to come out among the people that walk with their heads downward! The Antipathies, I think—’ (she was rather glad there was no one listening, this time, as it didn’t sound at all the right word) ‘—but I shall have to ask them what the name of the country is, you know. Please, Ma’am, is this New Zealand or Australia?’ (and she tried to curtsey as she spoke—fancy curtseying as you’re falling through the air! Do you think you could manage it?) ‘And what an ignorant little girl she’ll think me for asking! No, it’ll never do to ask: perhaps I shall see it written up somewhere.’

Down, down, down. There was nothing else to do, so Alice soon began talking again. ‘Dinah’ll miss me very much to-night, I should think!’ (Dinah was the cat.) ‘I hope they’ll remember her saucer of milk at tea-time. Dinah my dear! I wish you were down here with me! There are no mice in the air, I’m afraid, but you might catch a bat, and that’s very like a mouse, you know. But do cats eat bats, I wonder?’ And here Alice began to get rather sleepy, and went on saying to herself, in a dreamy sort of way, ‘Do cats eat bats? Do cats eat bats?’ and sometimes, ‘Do bats eat cats?’ for, you see, as she couldn’t answer either question, it didn’t much matter which way she put it. She felt that she was dozing off, and had just begun to dream that she was walking hand in hand with Dinah, and saying to her very earnestly, ‘Now, Dinah, tell me the truth: did you ever eat a bat?’ when suddenly, thump! thump! down she came upon a heap of sticks and dry leaves, and the fall was over.

Alice was not a bit hurt, and she jumped up on to her feet in a moment: she looked up, but it was all dark overhead; before her was another long passage, and the White Rabbit was still in sight, hurrying down it. There was not a moment to be lost: away went Alice like the wind, and was just in time to hear it say, as it turned a corner, ‘Oh my ears and whiskers, how late it’s getting!’ She was close behind it when she turned the corner, but the Rabbit was no longer to be seen: she found herself in a long, low hall, which was lit up by a row of lamps hanging from the roof.

There were doors all round the hall, but they were all locked; and when Alice had been all the way down one side and up the other, trying every door, she walked sadly down the middle, wondering how she was ever to get out again.

Suddenly she came upon a little three-legged table, all made of solid glass; there was nothing on it except a tiny golden key, and Alice’s first thought was that it might belong to one of the doors of the hall; but, alas! either the locks were too large, or the key was too small, but at any rate it would not open any of them. However, on the second time round, she came upon a low curtain she had not noticed before, and behind it was a little door about fifteen inches high: she tried the little golden key in the lock, and to her great delight it fitted!

Alice opened the door and found that it led into a small passage, not much larger than a rat-hole: she knelt down and looked along the passage into the loveliest garden you ever saw. How she longed to get out of that dark hall, and wander about among those beds of bright flowers and those cool fountains, but she could not even get her head through the doorway; ‘and even if my head would go through,’ thought poor Alice, ‘it would be of very little use without my shoulders. Oh, how I wish I could shut up like a telescope! I think I could, if I only knew how to begin.’ For, you see, so many out-of-the-way things had happened lately, that Alice had begun to think that very few things indeed were really impossible.

There seemed to be no use in waiting by the little door, so she went back to the table, half hoping she might find another key on it, or at any rate a book of rules for shutting people up like telescopes: this time she found a little bottle on it, (‘which certainly was not here before,’ said Alice,) and round the neck of the bottle was a paper label, with the words ‘DRINK ME’ beautifully printed on it in large letters.

It was all very well to say ‘Drink me,’ but the wise little Alice was not going to do that in a hurry. ‘No, I’ll look first,’ she said, ‘and see whether it’s marked “poison” or not’; for she had read several nice little histories about children who had got burnt, and eaten up by wild beasts and other unpleasant things, all because they would not remember the simple rules their friends had taught them: such as, that a red-hot poker will burn you if you hold it too long; and that if you cut your finger very deeply with a knife, it usually bleeds; and she had never forgotten that, if you drink much from a bottle marked ‘poison,’ it is almost certain to disagree with you, sooner or later.

However, this bottle was not marked ‘poison,’ so Alice ventured to taste it, and finding it very nice, (it had, in fact, a sort of mixed flavour of cherry-tart, custard, pine-apple, roast turkey, toffee, and hot buttered toast,) she very soon finished it off.

  *    *    *    *    *    *    *

    *    *    *    *    *    *

  *    *    *    *    *    *    *

‘What a curious feeling!’ said Alice; ‘I must be shutting up like a telescope.’

And so it was indeed: she was now only ten inches high, and her face brightened up at the thought that she was now the right size for going through the little door into that lovely garden. First, however, she waited for a few minutes to see if she was going to shrink any further: she felt a little nervous about this; ‘for it might end, you know,’ said Alice to herself, ‘in my going out altogether, like a candle. I wonder what I should be like then?’ And she tried to fancy what the flame of a candle is like after the candle is blown out, for she could not remember ever having seen such a thing.

After a while, finding that nothing more happened, she decided on going into the garden at once; but, alas for poor Alice! when she got to the door, she found she had forgotten the little golden key, and when she went back to the table for it, she found she could not possibly reach it: she could see it quite plainly through the glass, and she tried her best to climb up one of the legs of the table, but it was too slippery; and when she had tired herself out with trying, the poor little thing sat down and cried.

‘Come, there’s no use in crying like that!’ said Alice to herself, rather sharply; ‘I advise you to leave off this minute!’ She generally gave herself very good advice, (though she very seldom followed it), and sometimes she scolded herself so severely as to bring tears into her eyes; and once she remembered trying to box her own ears for having cheated herself in a game of croquet she was playing against herself, for this curious child was very fond of pretending to be two people. ‘But it’s no use now,’ thought poor Alice, ‘to pretend to be two people! Why, there’s hardly enough of me left to make one respectable person!’

Soon her eye fell on a little glass box that was lying under the table: she opened it, and found in it a very small cake, on which the words ‘EAT ME’ were beautifully marked in currants. ‘Well, I’ll eat it,’ said Alice, ‘and if it makes me grow larger, I can reach the key; and if it makes me grow smaller, I can creep under the door; so either way I’ll get into the garden, and I don’t care which happens!’

She ate a little bit, and said anxiously to herself, ‘Which way? Which way?’, holding her hand on the top of her head to feel which way it was growing, and she was quite surprised to find that she remained the same size: to be sure, this generally happens when one eats cake, but Alice had got so much into the way of expecting nothing but out-of-the-way things to happen, that it seemed quite dull and stupid for life to go on in the common way.

So she set to work, and very soon finished off the cake. 
`

func BenchmarkNaiveSubstr(b *testing.B) {
	var s string
	for i := 0; i < b.N; i++ {
		s = a[5:10]
	}
	_ = s
}

func BenchmarkRopeSubstr(b *testing.B) {
	var s string
	r := New()
	r.Insert(0, a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = r.Substr(5, 10)
	}
	_ = s
}

func BenchmarkNaiveAppendAndDelete(b *testing.B) {
	const c = "ADDED"
	buf := bytes.NewBuffer([]byte(a))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(buf, c)
		buf.Truncate(len(c))
	}
}

func BenchmarkRopeAppendAndDelete(b *testing.B) {
	const c = "ADDED"
	r := New()
	r.Insert(0, a)

	// these are expensive operations because we're not working with byte slices
	la := len([]rune(a))
	lc := len([]rune(c))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Insert(la, c)
		r.EraseAt(la, lc)
	}
}

func BenchmarkNaiveAppend(b *testing.B) {
	buf := bytes.NewBuffer([]byte(a))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(buf, "a")
	}
}

func BenchmarkRopeAppend(b *testing.B) {
	r := New()
	r.Insert(0, a)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Insert(len(a), "a")
	}
}

func BenchmarkNaiveRandomInsert(b *testing.B) {
	buf := bytes.NewBuffer([]byte(a))
	b.ResetTimer()
	rand := 4
	for i := 0; i < b.N; i++ {
		if i%100 == 0 {
			rand = 2
		}
		buf.Grow(1)
		bb := buf.Bytes()
		// insert at point 4. Confirmed random
		copy(bb[rand+1:], bb[rand:])
		bb[rand] = 'a'
	}
}

func BenchmarkRopeRandomInsert(b *testing.B) {
	r := New()
	r.Insert(0, a)

	b.ResetTimer()
	rand := 4
	for i := 0; i < b.N; i++ {
		if i%100 == 0 {
			rand = 2
		}
		r.InsertRunes(rand, []rune{'a'})
	}
}
