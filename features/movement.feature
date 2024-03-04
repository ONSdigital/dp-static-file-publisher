Feature: Move files trigger by v2 messages

  Scenario: The one where there is single chunk file to move
    Given there is a single chunk file "data/single-chunk.txt" in the private bucket with content:
    """
    This is the file content for testing
    """
    When a message to publish the file "data/single-chunk.txt" is sent
    Then the public bucket contains a moved file called "data/single-chunk.txt"
    And the content of file "data/single-chunk.txt" in the public bucket matches the original plain text content
    And the private bucket still has a file called "data/single-chunk.txt"
    And the files API should be informed the file "data/single-chunk.txt" has been moved

  Scenario: The one where there is multi-chunk file to move
    Given there is a multi-chunk file "data/multi-chunk.txt" in the private bucket
    When a message to publish the file "data/multi-chunk.txt" is sent
    Then the public bucket contains a moved file called "data/multi-chunk.txt"
    And the content of file "data/multi-chunk.txt" in the public bucket matches the original plain text content
    And the private bucket still has a file called "data/multi-chunk.txt"
    And the files API should be informed the file "data/multi-chunk.txt" has been moved