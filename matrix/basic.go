// Copyright 2012 Harry de Boer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

// Zeros returns a zero-filled m x n matrix.
func Zeros(m, n int) *Matrix {
	return &Matrix{m, n, n, make([]float64, m * n)}
}

// Ones returns a one-filled m x n matrix.
func Ones(m, n int) *Matrix {
	A := Zeros(m, n)
	for i := range(A.data) {
		A.data[i] = 1.0
	}
	return A
}

// Identity returns an n x n identity matrix.
func Identity(n int) *Matrix {
	A := Zeros(n, n)

	for i := 0; i < len(A.data); i += (n + 1) {
		A.data[i] = 1.0
	}

	return A
}

// New returns a new m x n matrix with the specified contents.
func New(m, n int, data []float64) *Matrix {
	if len(data) != m * n {
		panic("matrix.New: length of the data does not match the specified dimensions.")
	}
	return &Matrix{m, n, n, data}
}

// Copy the contents of B to A.
func (A *Matrix) Copy(B *Matrix) {

	// Normal matrices.
	if B.stride == B.width {
		for i, bi := range B.data {
			A.data[i] = bi
		}
		return
	}

	// Submatrices.
	for i := 0; i < A.height; i++ {
		Ai := A.Row(i)
		for j, bij := range B.Row(i) {
			Ai[j] = bij
		}
	}
}

// Submatrix returns the m x n matrix that starts at row i and column j.
// The returned matrix shares its data with the original.
func (A *Matrix) SubMatrix(i, j, m, n int) *Matrix {
	return &Matrix{m, n, A.stride, A.data[i * A.stride + j: (i + m - 1) * A.stride + (j + n)]}
}

// At returns the value of the matrix at position (x, y)
func (A *Matrix) At(x, y int) float64 {
	return A.data[y * A.stride + x]
}

// Set changes the value of the matrix at position (x, y).
func (A *Matrix) Set(x, y int, v float64){
	A.data[y * A.stride + x] = v
}

// Rows returns the number of rows.
func (A *Matrix) Rows() int {
	return A.height
}

// Row returns the ith row.
func (A *Matrix) Row(i int) []float64 {
	return A.data[i * A.stride : i * A.stride + A.width]
}

// Cols returns the number of rows.
func (A *Matrix) Cols() int {
	return A.width
}

// RowSlices returns the contents of the matrix as a list of row vectors.
func (A *Matrix) RowVectors() [][]float64 {
	rows := make([][]float64, A.height)
	offset := 0

	for i := range rows {
		rows[i] = A.data[offset : offset + A.width]
		offset = offset + A.stride
	}

	return rows
}

// Transpose return the matrix transpose of A.
func Transpose(A *Matrix) *Matrix {
	B := Zeros(A.width, A.height)
	offset := 0

	for i := range B.data {
		B.data[i] = A.data[offset]

		offset += A.stride
		if offset >= len(A.data) {
			offset = offset % len(A.data) + 1
		}
	}

	return B
}
