// internal/bc/contracts/DocumentToken.sol
// SPDX-License-Identifier: MIT
pragma solidity >0.8.0 <0.9.0;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Burnable.sol";

contract DocumentToken is ERC721 {
    uint256 private _tokenIdCounter;

    constructor() ERC721("DocumentToken", "DOCTKN") {}

    struct Document {
        string docId;
        string docHash;
        string ownerEmailIdHash;
        uint256 uploadedAt;
    }

    mapping(uint256 => Document) private _documents;

    // Event to emit when a document is minted
    event DocumentMinted(
        uint256 tokenId,
        string docId,
        string docHash,
        string ownerEmailIdHash,
        uint256 uploadedAt
    );


    function mintDocument(
        string memory _docId,
        string memory _docHash,
        string memory _ownerEmailIdHash
    ) public {
        uint256 newTokenId = _tokenIdCounter;

        _documents[newTokenId] = Document({
            docId: _docId,
            docHash: _docHash,
            ownerEmailIdHash: _ownerEmailIdHash,
            uploadedAt: block.timestamp
        });


        _tokenIdCounter += 1;

        // Emit the DocumentMinted event
        emit DocumentMinted(
            newTokenId,
            _docId,
            _docHash,
            _ownerEmailIdHash,
            block.timestamp
        );

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
