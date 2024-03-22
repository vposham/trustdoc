// internal/bc/contracts/DocumentToken.sol
// SPDX-License-Identifier: MIT
pragma solidity >0.8.0 < 0.9.0;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

contract DocumentToken is ERC721 {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;

    struct Document {
        string docId;
        string docHash;
        string ownerEmailIdHash;
        uint256 uploadedAt;
    }

    mapping(uint256 => Document) private _documents;

    // Event to emit when a document is minted
    event DocumentMinted(
        uint256 indexed tokenId,
        string docId,
        string docHash,
        string ownerEmailIdHash,
        uint256 uploadedAt
    );

    constructor() ERC721("DocumentToken", "DOCTKN") {}

    function mintDocument(
        string memory _docId,
        string memory _docHash,
        string memory _ownerEmailIdHash
    ) public returns (uint256) {
        _tokenIds.increment();
        uint256 newItemId = _tokenIds.current();

        _documents[newItemId] = Document({
            docId: _docId,
            docHash: _docHash,
            ownerEmailIdHash: _ownerEmailIdHash,
            uploadedAt: block.timestamp
        });

        _mint(msg.sender, newItemId);

        // Emit the DocumentMinted event
        emit DocumentMinted(
            newItemId,
            _docId,
            _docHash,
            _ownerEmailIdHash,
            block.timestamp
        );

        return newItemId;
    }

    function getDocument(uint256 _tokenId) public view returns (string memory, string memory, string memory, uint256) {
        Document storage doc = _documents[_tokenId];
        return (doc.docId, doc.docHash, doc.ownerEmailIdHash, doc.uploadedAt);
    }

    function getDocumentContent(uint256 _tokenId) public view returns (string memory) {
        Document storage doc = _documents[_tokenId];
        return doc.docHash;
    }

    function getDocumentOwner(uint256 _tokenId) public view returns (string memory) {
        Document storage doc = _documents[_tokenId];
        return doc.ownerEmailIdHash;
    }
}
