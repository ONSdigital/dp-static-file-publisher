Feature: Decrypt files trigger by v2 messages

  Scenario: The one where there is single chunk file to decrypt
    Given there is a encrypted single chunk file "data/single-chunk.txt" in the private bucket with content:
    """
    This is the file content for testing
    """
    And there is an encryption key for file "data/single-chunk.txt" in vault
    And files API has file "data/single-chunk.txt" registered as published
    When a message to publish the file "data/single-chunk.txt" is sent
    Then the public bucket contains a decrypted file called "data/single-chunk.txt"
    And the content of file "data/single-chunk.txt" in the public bucket matches the original plain text content
    And the private bucket still has a encrypted file called "data/single-chunk.txt"
    And the files API should be informed the file has been decrypted

#  Scenario: The one where there is multi chunk file to decrypt
#    Given there is a encrypted file "data/multi-chunk.txt" in the private bucket
#    And there is an encryption key for file "data/multi-chunk.txt" in vault
#    When a message to publish the file "data/multi-chunk.txt" is sent
#    Then the public bucket contains a decrypted file called "data/multi-chunk.txt"
#    And the content of file "data/multi-chunk.txt" in the public bucket matches the original plain text content
#    And the private bucket still has a encrypted file called "data/multi-chunk.txt"