UPDATE roadmap_items
SET specifications = (
  SELECT json_group_array(json_object('title', value, 'description', '', 'done', 0))
  FROM json_each(roadmap_items.specifications)
)
WHERE json_valid(specifications)
  AND json_type(specifications) = 'array'
  AND json_array_length(specifications) > 0
  AND json_type(specifications, '$[0]') = 'text';
