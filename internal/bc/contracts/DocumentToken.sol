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
        string ownerName;
        uint256 uploadedAt;
    }

    mapping(uint256 => Document) private _documents;

    constructor() ERC721("DocumentToken", "DOCTKN") {}

    function mintDocument(
        string memory _docId,
        string memory _docHash,
        string memory _ownerName
    ) public returns (uint256) {
        _tokenIds.increment();
        uint256 newItemId = _tokenIds.current();

        _documents[newItemId] = Document({
            docId: _docId,
            docHash: _docHash,
            ownerName: _ownerName,
            uploadedAt: block.timestamp
        });

        _mint(msg.sender, newItemId);

        return newItemId;
    }

    function verifyDocument(
        uint256 _tokenId,
        string memory _docId,
        string memory _docHash,
        string memory _ownerName
    ) public view returns (bool) {
        Document storage doc = _documents[_tokenId];
        return (
            keccak256(abi.encodePacked(doc.docId)) == keccak256(abi.encodePacked(_docId)) &&
            keccak256(abi.encodePacked(doc.docHash)) == keccak256(abi.encodePacked(_docHash)) &&
            keccak256(abi.encodePacked(doc.ownerName)) == keccak256(abi.encodePacked(_ownerName))
        );
    }

    function getDocumentUploadedAt(uint256 _tokenId) public view returns (uint256) {
        return _documents[_tokenId].uploadedAt;
    }
}
