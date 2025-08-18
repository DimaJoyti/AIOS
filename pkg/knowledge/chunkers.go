package knowledge

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FixedSizeChunker implements fixed-size chunking
type FixedSizeChunker struct{}

// Chunk splits text into fixed-size chunks
func (c *FixedSizeChunker) Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error) {
	var chunks []*DocumentChunk

	if len(text) <= chunkSize {
		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     text,
			ChunkIndex:  0,
			StartOffset: 0,
			EndOffset:   len(text),
			CreatedAt:   time.Now(),
		}
		return []*DocumentChunk{chunk}, nil
	}

	for i := 0; i < len(text); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     text[i:end],
			ChunkIndex:  len(chunks),
			StartOffset: i,
			EndOffset:   end,
			CreatedAt:   time.Now(),
		}
		chunks = append(chunks, chunk)

		if end >= len(text) {
			break
		}
	}

	return chunks, nil
}

// SentenceChunker implements sentence-based chunking
type SentenceChunker struct{}

// Chunk splits text into sentence-based chunks
func (c *SentenceChunker) Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error) {
	// Split text into sentences
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRegex.Split(text, -1)

	var chunks []*DocumentChunk
	var currentChunk strings.Builder
	var currentSentences []string
	currentOffset := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Check if adding this sentence would exceed chunk size
		testContent := currentChunk.String()
		if testContent != "" {
			testContent += ". " + sentence
		} else {
			testContent = sentence
		}

		if len(testContent) > chunkSize && currentChunk.Len() > 0 {
			// Create chunk from current sentences
			chunk := &DocumentChunk{
				ID:          uuid.New().String(),
				Content:     currentChunk.String(),
				ChunkIndex:  len(chunks),
				StartOffset: currentOffset,
				EndOffset:   currentOffset + currentChunk.Len(),
				CreatedAt:   time.Now(),
			}
			chunks = append(chunks, chunk)

			// Handle overlap
			overlapSentences := 0
			if overlap > 0 && len(currentSentences) > 1 {
				overlapSentences = min(len(currentSentences)-1, overlap/100) // Rough overlap calculation
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			currentSentences = currentSentences[len(currentSentences)-overlapSentences:]
			currentOffset = chunk.EndOffset - calculateOverlapLength(currentSentences)

			for j, overlapSent := range currentSentences {
				if j > 0 {
					currentChunk.WriteString(". ")
				}
				currentChunk.WriteString(overlapSent)
			}

			if len(currentSentences) > 0 {
				currentChunk.WriteString(". ")
			}
		}

		// Add current sentence
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(". ")
		}
		currentChunk.WriteString(sentence)
		currentSentences = append(currentSentences, sentence)
	}

	// Add final chunk if there's remaining content
	if currentChunk.Len() > 0 {
		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     currentChunk.String(),
			ChunkIndex:  len(chunks),
			StartOffset: currentOffset,
			EndOffset:   currentOffset + currentChunk.Len(),
			CreatedAt:   time.Now(),
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// ParagraphChunker implements paragraph-based chunking
type ParagraphChunker struct{}

// Chunk splits text into paragraph-based chunks
func (c *ParagraphChunker) Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error) {
	// Split text into paragraphs
	paragraphs := strings.Split(text, "\n\n")

	var chunks []*DocumentChunk
	var currentChunk strings.Builder
	var currentParagraphs []string
	currentOffset := 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// Check if adding this paragraph would exceed chunk size
		testContent := currentChunk.String()
		if testContent != "" {
			testContent += "\n\n" + paragraph
		} else {
			testContent = paragraph
		}

		if len(testContent) > chunkSize && currentChunk.Len() > 0 {
			// Create chunk from current paragraphs
			chunk := &DocumentChunk{
				ID:          uuid.New().String(),
				Content:     currentChunk.String(),
				ChunkIndex:  len(chunks),
				StartOffset: currentOffset,
				EndOffset:   currentOffset + currentChunk.Len(),
				CreatedAt:   time.Now(),
			}
			chunks = append(chunks, chunk)

			// Handle overlap
			overlapParagraphs := 0
			if overlap > 0 && len(currentParagraphs) > 1 {
				overlapParagraphs = 1 // Usually one paragraph overlap
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			currentParagraphs = currentParagraphs[len(currentParagraphs)-overlapParagraphs:]
			currentOffset = chunk.EndOffset - calculateOverlapLength(currentParagraphs)

			for j, overlapPara := range currentParagraphs {
				if j > 0 {
					currentChunk.WriteString("\n\n")
				}
				currentChunk.WriteString(overlapPara)
			}

			if len(currentParagraphs) > 0 {
				currentChunk.WriteString("\n\n")
			}
		}

		// Add current paragraph
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
		currentParagraphs = append(currentParagraphs, paragraph)
	}

	// Add final chunk if there's remaining content
	if currentChunk.Len() > 0 {
		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     currentChunk.String(),
			ChunkIndex:  len(chunks),
			StartOffset: currentOffset,
			EndOffset:   currentOffset + currentChunk.Len(),
			CreatedAt:   time.Now(),
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// RecursiveChunker implements recursive character-based chunking
type RecursiveChunker struct{}

// Chunk splits text using recursive character splitting
func (c *RecursiveChunker) Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error) {
	// Define separators in order of preference
	separators := []string{"\n\n", "\n", ". ", " ", ""}

	return c.recursiveChunk(text, chunkSize, overlap, separators, 0)
}

// recursiveChunk performs recursive chunking
func (c *RecursiveChunker) recursiveChunk(text string, chunkSize int, overlap int, separators []string, offset int) ([]*DocumentChunk, error) {
	if len(text) <= chunkSize {
		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     text,
			ChunkIndex:  0,
			StartOffset: offset,
			EndOffset:   offset + len(text),
			CreatedAt:   time.Now(),
		}
		return []*DocumentChunk{chunk}, nil
	}

	var chunks []*DocumentChunk

	// Try each separator
	for _, separator := range separators {
		if separator == "" {
			// Last resort: character-level splitting
			return c.characterSplit(text, chunkSize, overlap, offset)
		}

		if strings.Contains(text, separator) {
			parts := strings.Split(text, separator)
			var currentChunk strings.Builder
			currentOffset := offset

			for _, part := range parts {
				testContent := currentChunk.String()
				if testContent != "" {
					testContent += separator + part
				} else {
					testContent = part
				}

				if len(testContent) > chunkSize && currentChunk.Len() > 0 {
					// Create chunk and recurse if needed
					chunkContent := currentChunk.String()
					if len(chunkContent) > chunkSize {
						// Recurse with remaining separators
						subChunks, err := c.recursiveChunk(chunkContent, chunkSize, overlap, separators[1:], currentOffset)
						if err != nil {
							return nil, err
						}
						chunks = append(chunks, subChunks...)
					} else {
						chunk := &DocumentChunk{
							ID:          uuid.New().String(),
							Content:     chunkContent,
							ChunkIndex:  len(chunks),
							StartOffset: currentOffset,
							EndOffset:   currentOffset + len(chunkContent),
							CreatedAt:   time.Now(),
						}
						chunks = append(chunks, chunk)
					}

					// Start new chunk with overlap
					currentOffset += len(chunkContent) + len(separator)
					currentChunk.Reset()

					// Add overlap if needed
					if overlap > 0 && len(chunkContent) > overlap {
						overlapText := chunkContent[len(chunkContent)-overlap:]
						currentChunk.WriteString(overlapText)
						currentOffset -= overlap
					}
				}

				// Add current part
				if currentChunk.Len() > 0 {
					currentChunk.WriteString(separator)
				}
				currentChunk.WriteString(part)
			}

			// Handle remaining content
			if currentChunk.Len() > 0 {
				chunkContent := currentChunk.String()
				if len(chunkContent) > chunkSize {
					// Recurse with remaining separators
					subChunks, err := c.recursiveChunk(chunkContent, chunkSize, overlap, separators[1:], currentOffset)
					if err != nil {
						return nil, err
					}
					chunks = append(chunks, subChunks...)
				} else {
					chunk := &DocumentChunk{
						ID:          uuid.New().String(),
						Content:     chunkContent,
						ChunkIndex:  len(chunks),
						StartOffset: currentOffset,
						EndOffset:   currentOffset + len(chunkContent),
						CreatedAt:   time.Now(),
					}
					chunks = append(chunks, chunk)
				}
			}

			return chunks, nil
		}
	}

	// Fallback to character splitting
	return c.characterSplit(text, chunkSize, overlap, offset)
}

// characterSplit performs character-level splitting
func (c *RecursiveChunker) characterSplit(text string, chunkSize int, overlap int, offset int) ([]*DocumentChunk, error) {
	var chunks []*DocumentChunk

	for i := 0; i < len(text); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := &DocumentChunk{
			ID:          uuid.New().String(),
			Content:     text[i:end],
			ChunkIndex:  len(chunks),
			StartOffset: offset + i,
			EndOffset:   offset + end,
			CreatedAt:   time.Now(),
		}
		chunks = append(chunks, chunk)

		if end >= len(text) {
			break
		}
	}

	return chunks, nil
}

// SemanticChunker implements semantic-based chunking
type SemanticChunker struct{}

// Chunk splits text based on semantic boundaries
func (c *SemanticChunker) Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error) {
	// For now, fall back to sentence chunking
	// In a real implementation, this would use semantic similarity
	// to determine optimal chunk boundaries
	sentenceChunker := &SentenceChunker{}
	return sentenceChunker.Chunk(text, chunkSize, overlap)
}

// Helper functions

// calculateOverlapLength calculates the total length of overlap content
func calculateOverlapLength(content []string) int {
	total := 0
	for i, item := range content {
		if i > 0 {
			total += 2 // Add separator length
		}
		total += len(item)
	}
	return total
}
